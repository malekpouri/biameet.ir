package services

import (
	"fmt"
	"time"

	"biameet.ir/db"
	"biameet.ir/models"
	"github.com/google/uuid"
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

		// Check for double vote (handled by UNIQUE constraint, but good to check or handle error)
		// We will try to insert and handle duplicate error, or just let it fail.
		// The requirement says "Prevent double-vote".
		// Let's use INSERT OR REPLACE or just INSERT and catch error?
		// Requirement: "Prevent double-vote (same name on same session+timeslot)"
		// My schema has UNIQUE(timeslot_id, voter_name).

		voteID := uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO votes (id, timeslot_id, voter_name, note, created_at_utc)
			VALUES (?, ?, ?, ?, ?)
		`, voteID, item.TimeslotID, req.VoterName, item.Note, createdAt)

		if err != nil {
			// Check if it's a constraint violation
			// For sqlite3, error message contains "UNIQUE constraint failed"
			if err.Error() == "UNIQUE constraint failed: votes.timeslot_id, votes.voter_name" ||
				(len(err.Error()) > 25 && err.Error()[:25] == "UNIQUE constraint failed") {
				return fmt.Errorf("user %s has already voted for timeslot %s", req.VoterName, item.TimeslotID)
			}
			return err
		}
	}

	return tx.Commit()
}
