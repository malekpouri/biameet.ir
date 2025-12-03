package services

	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"biameet.ir/db"
	"biameet.ir/models"
	"biameet.ir/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

func CreateSession(req models.CreateSessionRequest) (*models.CreateSessionResponse, error) {
	sessionID := utils.GenerateShortID(5)
	createdAt := time.Now().UTC().Format(time.RFC3339)

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Serialize DynamicConfig
	var dynamicConfigJSON string
	if req.DynamicConfig != nil {
		bytes, err := json.Marshal(req.DynamicConfig)
		if err != nil {
			return nil, err
		}
		dynamicConfigJSON = string(bytes)
	}

	// Default type if empty
	sessionType := req.Type
	if sessionType == "" {
		sessionType = "fixed"
	}

	// Insert Session
	_, err = tx.Exec(`
		INSERT INTO sessions (id, title, creator_name, created_at_utc, type, dynamic_config)
		VALUES (?, ?, ?, ?, ?, ?)
	`, sessionID, req.Title, req.CreatorName, createdAt, sessionType, dynamicConfigJSON)
	if err != nil {
		return nil, err
	}

	// Insert Timeslots
	for _, ts := range req.Timeslots {
		tsID := uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO timeslots (id, session_id, start_utc, end_utc)
			VALUES (?, ?, ?, ?)
		`, tsID, sessionID, ts.StartUTC, ts.EndUTC)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &models.CreateSessionResponse{
		ID:   sessionID,
		Link: "/sessions/" + sessionID, // Frontend route
	}, nil
}

func AddTimeslot(sessionID string, req models.TimeslotRequest) (*models.Timeslot, error) {
	// Check for duplicates
	var count int
	err := db.DB.QueryRow(`
		SELECT COUNT(*) FROM timeslots 
		WHERE session_id = ? AND start_utc = ? AND end_utc = ?
	`, sessionID, req.StartUTC, req.EndUTC).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("این زمان قبلاً ثبت شده است")
	}

	tsID := uuid.New().String()

	var passwordHash sql.NullString
	if req.Password != "" {
		bytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		passwordHash.String = string(bytes)
		passwordHash.Valid = true
	}

	_, err = db.DB.Exec(`
		INSERT INTO timeslots (id, session_id, start_utc, end_utc, created_by, password_hash)
		VALUES (?, ?, ?, ?, ?, ?)
	`, tsID, sessionID, req.StartUTC, req.EndUTC, req.CreatedBy, passwordHash)
	if err != nil {
		return nil, err
	}

	return &models.Timeslot{
		ID:        tsID,
		SessionID: sessionID,
		StartUTC:  req.StartUTC,
		EndUTC:    req.EndUTC,
		CreatedBy: req.CreatedBy,
	}, nil
}

func DeleteTimeslot(sessionID, timeslotID, password string) error {
	// Check if timeslot exists and belongs to session
	var count int
	var storedHash sql.NullString
	err := db.DB.QueryRow("SELECT COUNT(*), password_hash FROM timeslots WHERE id = ? AND session_id = ?", timeslotID, sessionID).Scan(&count, &storedHash)
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("timeslot not found")
	}

	// Check if timeslot has votes
	var voteCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM votes WHERE timeslot_id = ?", timeslotID).Scan(&voteCount)
	if err != nil {
		return err
	}
	if voteCount > 0 {
		return fmt.Errorf("cannot delete timeslot with existing votes")
	}

	// Check password if set
	if storedHash.Valid && storedHash.String != "" {
		if password == "" {
			return fmt.Errorf("password_required")
		}
		err = bcrypt.CompareHashAndPassword([]byte(storedHash.String), []byte(password))
		if err != nil {
			return fmt.Errorf("invalid_password")
		}
	}

	_, err = db.DB.Exec("DELETE FROM timeslots WHERE id = ?", timeslotID)
	return err
}
