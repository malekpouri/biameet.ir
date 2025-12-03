package services

import (
	"database/sql"
	"fmt"
	"time"

	"biameet.ir/db"
	"biameet.ir/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func SubmitVote(sessionID string, req models.VoteRequest) error {
	// 1. Validate Session exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM sessions WHERE id = ?)", sessionID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("session not found")
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	createdAt := time.Now().UTC().Format(time.RFC3339)

	// 2. Handle Participant Logic
	var storedHash sql.NullString
	err = tx.QueryRow("SELECT password_hash FROM participants WHERE session_id = ? AND name = ?", sessionID, req.VoterName).Scan(&storedHash)

	if err == sql.ErrNoRows {
		// New participant
		var hash sql.NullString
		if req.Password != "" {
			bytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			hash.String = string(bytes)
			hash.Valid = true
		}

		_, err = tx.Exec("INSERT INTO participants (session_id, name, password_hash, created_at_utc) VALUES (?, ?, ?, ?)",
			sessionID, req.VoterName, hash, createdAt)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		// Existing participant
		if storedHash.Valid && storedHash.String != "" {
			// Password required
			if req.Password == "" {
				return fmt.Errorf("password_required") // Specific error for frontend to handle
			}
			err = bcrypt.CompareHashAndPassword([]byte(storedHash.String), []byte(req.Password))
			if err != nil {
				return fmt.Errorf("invalid_password") // Specific error
			}
		} else {
			// User exists but has no password set.
			// Prevent editing to avoid impersonation.
			return fmt.Errorf("name_taken_no_password")
		}

		// Delete existing votes for this user in this session
		_, err = tx.Exec(`
			DELETE FROM votes 
			WHERE voter_name = ? 
			AND timeslot_id IN (SELECT id FROM timeslots WHERE session_id = ?)
		`, req.VoterName, sessionID)
		if err != nil {
			return err
		}
	}

	// 3. Insert New Votes
	for _, item := range req.Votes {
		// Validate Timeslot belongs to Session
		var tsSessionID string
		err := tx.QueryRow("SELECT session_id FROM timeslots WHERE id = ?", item.TimeslotID).Scan(&tsSessionID)
		if err != nil {
			return fmt.Errorf("invalid timeslot id: %s", item.TimeslotID)
		}
		if tsSessionID != sessionID {
			return fmt.Errorf("timeslot %s does not belong to session %s", item.TimeslotID, sessionID)
		}

		voteID := uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO votes (id, timeslot_id, voter_name, note, created_at_utc)
			VALUES (?, ?, ?, ?, ?)
		`, voteID, item.TimeslotID, req.VoterName, item.Note, createdAt)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
