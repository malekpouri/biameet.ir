package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"biameet.ir/db"
	"biameet.ir/models"
	"biameet.ir/utils"
	"github.com/google/uuid"
)

// querier is satisfied by both *sql.DB and *sql.Tx.
type querier interface {
	QueryRow(query string, args ...any) *sql.Row
}

type sessionMeta struct {
	Type    string
	Config  *models.DynamicConfig
	Expired bool
}

func loadSessionMeta(q querier, id string) (*sessionMeta, error) {
	var typ, cfgJSON, expires sql.NullString
	err := q.QueryRow(`SELECT type, dynamic_config, expires_at_utc FROM sessions WHERE id = ?`, id).
		Scan(&typ, &cfgJSON, &expires)
	if err == sql.ErrNoRows {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	m := &sessionMeta{Type: typ.String, Expired: isExpired(expires.String)}
	if m.Type == "" {
		m.Type = "fixed"
	}
	if cfgJSON.String != "" {
		var cfg models.DynamicConfig
		if json.Unmarshal([]byte(cfgJSON.String), &cfg) == nil {
			m.Config = &cfg
		}
	}
	return m, nil
}

func isExpired(expiresAt string) bool {
	if expiresAt == "" {
		return false
	}
	t, err := time.Parse(time.RFC3339Nano, expiresAt)
	return err == nil && time.Now().After(t)
}

func participantHash(q querier, sessionID, name string) (hash sql.NullString, found bool, err error) {
	err = q.QueryRow(`SELECT password_hash FROM participants WHERE session_id = ? AND name = ?`, sessionID, name).Scan(&hash)
	if err == sql.ErrNoRows {
		return hash, false, nil
	}
	return hash, err == nil, err
}

// authorizeParticipant authenticates name for a write in sessionID. It runs
// outside any transaction because bcrypt is slow and must not hold the write
// lock. It returns the hash to store/keep and whether the participant is new.
func authorizeParticipant(sessionID, name, password, token string) (hash sql.NullString, isNew bool, err error) {
	stored, found, err := participantHash(db.DB, sessionID, name)
	if err != nil {
		return hash, false, err
	}
	if !found {
		hash, err = hashPassword(password)
		return hash, true, err
	}
	if !stored.Valid || stored.String == "" {
		// Registered without a password: nobody can prove they own this name.
		return hash, false, ErrNameTakenNoPassword
	}
	if err := checkCredential(sessionID, name, stored, password, token); err != nil {
		return hash, false, err
	}
	return stored, false, nil
}

// recheckParticipant makes sure the participant row didn't change between
// authorizeParticipant and the write transaction.
func recheckParticipant(tx *sql.Tx, sessionID, name string, hash sql.NullString, isNew bool) error {
	stored, found, err := participantHash(tx, sessionID, name)
	if err != nil {
		return err
	}
	if found == isNew || (found && stored.String != hash.String) {
		return ErrConcurrentUpdate
	}
	return nil
}

func CreateSession(req models.CreateSessionRequest) (*models.CreateSessionResponse, error) {
	if err := validateCreate(&req); err != nil {
		return nil, err
	}

	var cfgJSON, expires sql.NullString
	if req.DynamicConfig != nil {
		b, err := json.Marshal(req.DynamicConfig)
		if err != nil {
			return nil, err
		}
		cfgJSON = sql.NullString{String: string(b), Valid: true}
	}
	if req.ExpiresAtUTC != "" {
		expires = sql.NullString{String: req.ExpiresAtUTC, Valid: true}
	}
	createdAt := time.Now().UTC().Format(time.RFC3339)

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 62^5 IDs make collisions rare but not impossible; the write lock held by
	// this transaction makes check-then-insert safe.
	var sessionID string
	for attempt := 0; ; attempt++ {
		sessionID = utils.GenerateShortID(5)
		var taken bool
		if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM sessions WHERE id = ?)`, sessionID).Scan(&taken); err != nil {
			return nil, err
		}
		if !taken {
			break
		}
		if attempt == 10 {
			return nil, errors.New("could not allocate a session id")
		}
	}

	_, err = tx.Exec(`
		INSERT INTO sessions (id, title, creator_name, created_at_utc, expires_at_utc, type, dynamic_config)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, sessionID, req.Title, req.CreatorName, createdAt, expires, req.Type, cfgJSON)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool, len(req.Timeslots))
	for _, ts := range req.Timeslots {
		key := ts.StartUTC + "|" + ts.EndUTC
		if seen[key] {
			continue
		}
		seen[key] = true
		_, err = tx.Exec(`INSERT INTO timeslots (id, session_id, start_utc, end_utc) VALUES (?, ?, ?, ?)`,
			uuid.NewString(), sessionID, ts.StartUTC, ts.EndUTC)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &models.CreateSessionResponse{
		ID:   sessionID,
		Link: "/" + sessionID,
	}, nil
}

func AddTimeslot(sessionID string, req models.TimeslotRequest) (*models.AddTimeslotResponse, error) {
	var err error
	if req.CreatedBy, err = checkText(req.CreatedBy, "نام", maxNameLen, false); err != nil {
		return nil, err
	}
	if err := checkPassword(req.Password); err != nil {
		return nil, err
	}
	start, end, err := normalizeSlot(req.StartUTC, req.EndUTC)
	if err != nil {
		return nil, err
	}

	meta, err := loadSessionMeta(db.DB, sessionID)
	if err != nil {
		return nil, err
	}
	if meta.Expired {
		return nil, ErrSessionExpired
	}
	switch meta.Type {
	case "dynamic":
		if meta.Config == nil || !inDynamicRange(meta.Config, start, end) {
			return nil, ErrOutOfRange
		}
	case "weekly":
	default:
		return nil, ErrFixedSession
	}

	// The proposer's hash is stored on the timeslot as well, so deleting it
	// later requires the same credentials.
	var hash sql.NullString
	isNew := false
	if req.CreatedBy != "" {
		if hash, isNew, err = authorizeParticipant(sessionID, req.CreatedBy, req.Password, req.Token); err != nil {
			return nil, err
		}
	} else if hash, err = hashPassword(req.Password); err != nil {
		return nil, err
	}

	startS, endS := start.Format(isoLayout), end.Format(isoLayout)
	tsID := uuid.NewString()
	createdAt := time.Now().UTC().Format(time.RFC3339)

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if req.CreatedBy != "" {
		if err := recheckParticipant(tx, sessionID, req.CreatedBy, hash, isNew); err != nil {
			return nil, err
		}
	}

	var total, dup int
	err = tx.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(start_utc = ? AND end_utc = ?), 0)
		FROM timeslots WHERE session_id = ?
	`, startS, endS, sessionID).Scan(&total, &dup)
	if err != nil {
		return nil, err
	}
	if dup > 0 {
		return nil, ErrDuplicateTimeslot
	}
	if total >= maxSessionSlots {
		return nil, ErrTooManyTimeslots
	}

	createdBy := sql.NullString{String: req.CreatedBy, Valid: req.CreatedBy != ""}
	_, err = tx.Exec(`
		INSERT INTO timeslots (id, session_id, start_utc, end_utc, created_by, password_hash)
		VALUES (?, ?, ?, ?, ?, ?)
	`, tsID, sessionID, startS, endS, createdBy, hash)
	if err != nil {
		return nil, err
	}

	// The proposer automatically votes for their own slot.
	if req.CreatedBy != "" {
		if isNew {
			_, err = tx.Exec(`INSERT INTO participants (session_id, name, password_hash, created_at_utc) VALUES (?, ?, ?, ?)`,
				sessionID, req.CreatedBy, hash, createdAt)
			if err != nil {
				return nil, err
			}
		}
		_, err = tx.Exec(`INSERT INTO votes (id, timeslot_id, voter_name, created_at_utc) VALUES (?, ?, ?, ?)`,
			uuid.NewString(), tsID, req.CreatedBy, createdAt)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	resp := &models.AddTimeslotResponse{
		Timeslot: models.Timeslot{
			ID:        tsID,
			SessionID: sessionID,
			StartUTC:  startS,
			EndUTC:    endS,
			CreatedBy: req.CreatedBy,
			Votes:     []models.Vote{},
		},
	}
	if req.CreatedBy != "" {
		resp.Token = participantToken(sessionID, req.CreatedBy, hash.String)
	}
	return resp, nil
}

func DeleteTimeslot(sessionID, timeslotID string, req models.DeleteTimeslotRequest) error {
	meta, err := loadSessionMeta(db.DB, sessionID)
	if err != nil {
		return err
	}
	if meta.Expired {
		return ErrSessionExpired
	}

	var createdBy, hash sql.NullString
	err = db.DB.QueryRow(`SELECT created_by, password_hash FROM timeslots WHERE id = ? AND session_id = ?`,
		timeslotID, sessionID).Scan(&createdBy, &hash)
	if err == sql.ErrNoRows {
		return ErrTimeslotNotFound
	}
	if err != nil {
		return err
	}

	protected := hash.Valid && hash.String != ""
	if protected {
		authorized := false
		// A token proves identity as the proposer (tokens are bound to the participant's hash).
		if req.Token != "" && createdBy.String != "" {
			pHash, found, err := participantHash(db.DB, sessionID, createdBy.String)
			if err != nil {
				return err
			}
			authorized = found && pHash.String != "" &&
				checkCredential(sessionID, createdBy.String, pHash, "", req.Token) == nil
		}
		if !authorized {
			if err := checkCredential(sessionID, createdBy.String, hash, req.Password, ""); err != nil {
				return err
			}
		}
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Other people's votes always block deletion. The proposer's own automatic
	// vote doesn't when they've proven who they are.
	var blocking int
	if protected && createdBy.String != "" {
		err = tx.QueryRow(`SELECT COUNT(*) FROM votes WHERE timeslot_id = ? AND voter_name <> ?`, timeslotID, createdBy.String).Scan(&blocking)
	} else {
		err = tx.QueryRow(`SELECT COUNT(*) FROM votes WHERE timeslot_id = ?`, timeslotID).Scan(&blocking)
	}
	if err != nil {
		return err
	}
	if blocking > 0 {
		return ErrTimeslotHasVotes
	}

	if _, err = tx.Exec(`DELETE FROM votes WHERE timeslot_id = ?`, timeslotID); err != nil {
		return err
	}
	res, err := tx.Exec(`DELETE FROM timeslots WHERE id = ? AND session_id = ?`, timeslotID, sessionID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrTimeslotNotFound
	}
	return tx.Commit()
}
