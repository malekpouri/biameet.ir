package services

import (
	"time"

	"biameet.ir/db"
	"biameet.ir/models"
	"github.com/google/uuid"
)

func CreateSession(req models.CreateSessionRequest) (*models.CreateSessionResponse, error) {
	sessionID := uuid.New().String()
	createdAt := time.Now().UTC().Format(time.RFC3339)

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert Session
	_, err = tx.Exec(`
		INSERT INTO sessions (id, title, creator_name, created_at_utc)
		VALUES (?, ?, ?, ?)
	`, sessionID, req.Title, req.CreatorName, createdAt)
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
