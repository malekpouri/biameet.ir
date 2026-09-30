package services

import (
	"database/sql"
	"encoding/json"

	"biameet.ir/db"
	"biameet.ir/models"
)

func GetSession(id string) (*models.Session, error) {
	var session models.Session
	var expiresAt, archivedAt, dynamicConfigJSON, sessionType sql.NullString

	err := db.DB.QueryRow(`
		SELECT id, title, creator_name, created_at_utc, expires_at_utc, archived_at_utc, type, dynamic_config
		FROM sessions WHERE id = ?
	`, id).Scan(
		&session.ID, &session.Title, &session.CreatorName, &session.CreatedAtUTC,
		&expiresAt, &archivedAt, &sessionType, &dynamicConfigJSON,
	)
	if err == sql.ErrNoRows {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	session.ExpiresAtUTC = expiresAt.String
	session.ArchivedAtUTC = archivedAt.String
	session.Expired = isExpired(expiresAt.String)
	session.Type = sessionType.String
	if session.Type == "" {
		session.Type = "fixed"
	}
	if dynamicConfigJSON.String != "" {
		var config models.DynamicConfig
		if err := json.Unmarshal([]byte(dynamicConfigJSON.String), &config); err == nil {
			session.DynamicConfig = &config
		}
	}

	rows, err := db.DB.Query(`
		SELECT id, session_id, start_utc, end_utc, created_by
		FROM timeslots WHERE session_id = ?
		ORDER BY start_utc, end_utc
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	session.Timeslots = []models.Timeslot{}
	for rows.Next() {
		var ts models.Timeslot
		var createdBy sql.NullString
		if err := rows.Scan(&ts.ID, &ts.SessionID, &ts.StartUTC, &ts.EndUTC, &createdBy); err != nil {
			return nil, err
		}
		ts.CreatedBy = createdBy.String
		ts.Votes = []models.Vote{}
		session.Timeslots = append(session.Timeslots, ts)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	index := make(map[string]int, len(session.Timeslots))
	for i := range session.Timeslots {
		index[session.Timeslots[i].ID] = i
	}

	voteRows, err := db.DB.Query(`
		SELECT v.id, v.timeslot_id, v.voter_name, v.note, v.created_at_utc
		FROM votes v
		JOIN timeslots t ON v.timeslot_id = t.id
		WHERE t.session_id = ?
		ORDER BY v.created_at_utc
	`, id)
	if err != nil {
		return nil, err
	}
	defer voteRows.Close()

	for voteRows.Next() {
		var v models.Vote
		var note sql.NullString
		if err := voteRows.Scan(&v.ID, &v.TimeslotID, &v.VoterName, &note, &v.CreatedAtUTC); err != nil {
			return nil, err
		}
		v.Note = note.String
		if i, ok := index[v.TimeslotID]; ok {
			session.Timeslots[i].Votes = append(session.Timeslots[i].Votes, v)
		}
	}
	return &session, voteRows.Err()
}

// GetSessionSummary returns just what the HTML page needs for link previews,
// without loading timeslots and votes.
func GetSessionSummary(id string) (title, creator string, err error) {
	err = db.DB.QueryRow(`SELECT title, creator_name FROM sessions WHERE id = ?`, id).Scan(&title, &creator)
	if err == sql.ErrNoRows {
		err = ErrSessionNotFound
	}
	return
}
