package services

import (
	"database/sql"
	"fmt"

	"biameet.ir/db"
	"biameet.ir/models"
)

func GetSession(id string) (*models.Session, error) {
	var session models.Session

	var expiresAt, archivedAt sql.NullString

	// 1. Get Session
	err := db.DB.QueryRow(`
		SELECT id, title, creator_name, created_at_utc, expires_at_utc, archived_at_utc
		FROM sessions WHERE id = ?
	`, id).Scan(
		&session.ID, &session.Title, &session.CreatorName, &session.CreatedAtUTC,
		&expiresAt, &archivedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, err
	}

	if expiresAt.Valid {
		session.ExpiresAtUTC = expiresAt.String
	}
	if archivedAt.Valid {
		session.ArchivedAtUTC = archivedAt.String
	}

	// 2. Get Timeslots
	rows, err := db.DB.Query(`
		SELECT id, session_id, start_utc, end_utc
		FROM timeslots WHERE session_id = ?
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	timeslotMap := make(map[string]*models.Timeslot)
	var timeslots []models.Timeslot

	for rows.Next() {
		var ts models.Timeslot
		if err := rows.Scan(&ts.ID, &ts.SessionID, &ts.StartUTC, &ts.EndUTC); err != nil {
			return nil, err
		}
		ts.Votes = []models.Vote{} // Initialize empty slice
		timeslots = append(timeslots, ts)
		// We need to map pointers to modify the slice elements later?
		// Actually, appending to slice copies the struct.
		// Let's use index or pointers.
	}

	// Re-loop to create map for easy vote assignment
	// Or just query votes for all timeslots in one go if possible, or loop.
	// For simplicity, let's just loop and query votes (N+1 problem, but okay for small scale)
	// OR better: SELECT * FROM votes WHERE timeslot_id IN (...)

	// Let's do a slightly better approach: fetch all votes for these timeslots
	// But first, let's fix the slice/map issue.
	for i := range timeslots {
		timeslotMap[timeslots[i].ID] = &timeslots[i]
	}

	// 3. Get Votes
	// We can get all votes for the session via join or just get all votes for these timeslots
	// Since we don't have session_id in votes, we need to join or use IN clause.
	// Simple approach: Iterate timeslots (if few) or use IN.
	// Let's use a JOIN with timeslots to filter by session_id
	voteRows, err := db.DB.Query(`
		SELECT v.id, v.timeslot_id, v.voter_name, v.note, v.created_at_utc
		FROM votes v
		JOIN timeslots t ON v.timeslot_id = t.id
		WHERE t.session_id = ?
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
		if note.Valid {
			v.Note = note.String
		}
		if ts, ok := timeslotMap[v.TimeslotID]; ok {
			ts.Votes = append(ts.Votes, v)
		}
	}

	session.Timeslots = timeslots
	return &session, nil
}
