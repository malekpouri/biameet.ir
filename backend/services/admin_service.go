package services

import (
	"biameet.ir/db"
	"biameet.ir/models"
)

func GetAdminStats() (*models.AdminStats, error) {
	stats := &models.AdminStats{}

	// Total Sessions
	err := db.DB.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&stats.TotalSessions)
	if err != nil {
		return nil, err
	}

	// Total Timeslots
	err = db.DB.QueryRow("SELECT COUNT(*) FROM timeslots").Scan(&stats.TotalTimeslots)
	if err != nil {
		return nil, err
	}

	// Total Votes
	err = db.DB.QueryRow("SELECT COUNT(*) FROM votes").Scan(&stats.TotalVotes)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
