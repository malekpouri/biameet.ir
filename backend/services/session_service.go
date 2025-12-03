package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"biameet.ir/db"
	"biameet.ir/models"
	"biameet.ir/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

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

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO timeslots (id, session_id, start_utc, end_utc, created_by, password_hash)
		VALUES (?, ?, ?, ?, ?, ?)
	`, tsID, sessionID, req.StartUTC, req.EndUTC, req.CreatedBy, passwordHash)
	if err != nil {
		return nil, err
	}

	// Automatically vote for the creator if name is provided
	if req.CreatedBy != "" {
		voteID := uuid.New().String()
		createdAt := time.Now().UTC().Format(time.RFC3339)

		// We also need to register the participant if not exists, similar to SubmitVote logic.
		// However, reusing SubmitVote logic here is tricky because of circular dependencies or transaction handling.
		// Let's duplicate the minimal logic needed: Insert participant if new, then insert vote.

		// 1. Handle Participant
		var pStoredHash sql.NullString
		err = tx.QueryRow("SELECT password_hash FROM participants WHERE session_id = ? AND name = ?", sessionID, req.CreatedBy).Scan(&pStoredHash)
		if err == sql.ErrNoRows {
			// New participant
			_, err = tx.Exec("INSERT INTO participants (session_id, name, password_hash, created_at_utc) VALUES (?, ?, ?, ?)",
				sessionID, req.CreatedBy, passwordHash, createdAt) // Use same password hash for participant
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		} else {
			// Existing participant. We don't validate password here because they just created the timeslot.
			// But ideally we should? For now let's assume if they can create a timeslot, they can vote on it.
			// Actually, anyone can create a timeslot with any name.
			// If I use someone else's name to create a timeslot, I will also cast a vote for them?
			// This might be a security hole if I can impersonate.
			// But we added password check for editing votes.
			// If I create a timeslot as "Ali" (who has a password), and I don't provide password (or provide my own),
			// I shouldn't be able to register "Ali" as a participant if "Ali" already exists with a different password.

			if pStoredHash.Valid && pStoredHash.String != "" {
				// Existing user has password.
				if req.Password == "" {
					// If creator didn't provide password, they can't vote as this user.
					// But they just created the timeslot!
					// Let's just skip auto-voting if authentication fails, or error out?
					// Erroring out seems safer to prevent confusion.
					return nil, fmt.Errorf("password_required")
				}
				if err := bcrypt.CompareHashAndPassword([]byte(pStoredHash.String), []byte(req.Password)); err != nil {
					return nil, fmt.Errorf("invalid_password")
				}
			}
		}

		// 2. Insert Vote
		_, err = tx.Exec(`
			INSERT INTO votes (id, timeslot_id, voter_name, created_at_utc)
			VALUES (?, ?, ?, ?)
		`, voteID, tsID, req.CreatedBy, createdAt)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
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
