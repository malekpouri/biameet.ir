package services

import (
	"time"

	"biameet.ir/db"
	"biameet.ir/models"
	"github.com/google/uuid"
)

// SubmitVote replaces the voter's votes in a session. An empty vote list
// withdraws all of them. It returns a token the client can use to edit later.
func SubmitVote(sessionID string, req models.VoteRequest) (string, error) {
	var err error
	if req.VoterName, err = checkText(req.VoterName, "نام", maxNameLen, true); err != nil {
		return "", err
	}
	if err := checkPassword(req.Password); err != nil {
		return "", err
	}
	if len(req.Votes) > maxSessionSlots {
		return "", invalid("تعداد رای‌ها بیش از حد مجاز است")
	}
	for i := range req.Votes {
		if req.Votes[i].Note, err = checkText(req.Votes[i].Note, "یادداشت", maxNoteLen, false); err != nil {
			return "", err
		}
	}

	meta, err := loadSessionMeta(db.DB, sessionID)
	if err != nil {
		return "", err
	}
	if meta.Expired {
		return "", ErrSessionExpired
	}

	hash, isNew, err := authorizeParticipant(sessionID, req.VoterName, req.Password, req.Token)
	if err != nil {
		return "", err
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	if err := recheckParticipant(tx, sessionID, req.VoterName, hash, isNew); err != nil {
		return "", err
	}

	valid := map[string]bool{}
	rows, err := tx.Query(`SELECT id FROM timeslots WHERE session_id = ?`, sessionID)
	if err != nil {
		return "", err
	}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return "", err
		}
		valid[id] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return "", err
	}

	createdAt := time.Now().UTC().Format(time.RFC3339)

	if isNew {
		_, err = tx.Exec(`INSERT INTO participants (session_id, name, password_hash, created_at_utc) VALUES (?, ?, ?, ?)`,
			sessionID, req.VoterName, hash, createdAt)
	} else {
		_, err = tx.Exec(`
			DELETE FROM votes
			WHERE voter_name = ?
			AND timeslot_id IN (SELECT id FROM timeslots WHERE session_id = ?)
		`, req.VoterName, sessionID)
	}
	if err != nil {
		return "", err
	}

	seen := map[string]bool{}
	for _, item := range req.Votes {
		if !valid[item.TimeslotID] {
			return "", invalid("زمان انتخاب‌شده متعلق به این جلسه نیست یا حذف شده است")
		}
		if seen[item.TimeslotID] {
			continue
		}
		seen[item.TimeslotID] = true
		_, err = tx.Exec(`INSERT INTO votes (id, timeslot_id, voter_name, note, created_at_utc) VALUES (?, ?, ?, ?, ?)`,
			uuid.NewString(), item.TimeslotID, req.VoterName, item.Note, createdAt)
		if err != nil {
			return "", err
		}
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return participantToken(sessionID, req.VoterName, hash.String), nil
}
