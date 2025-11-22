package services

import (
	"encoding/json"
	"time"

	"biameet.ir/db"
	"biameet.ir/models"
	"biameet.ir/utils"
	"github.com/google/uuid"
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
	tsID := uuid.New().String()
	
	_, err := db.DB.Exec(`
		INSERT INTO timeslots (id, session_id, start_utc, end_utc)
		VALUES (?, ?, ?, ?)
	`, tsID, sessionID, req.StartUTC, req.EndUTC)
	if err != nil {
		return nil, err
	}

	return &models.Timeslot{
		ID:        tsID,
		SessionID: sessionID,
		StartUTC:  req.StartUTC,
		EndUTC:    req.EndUTC,
	}, nil
}
