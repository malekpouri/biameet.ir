package services

import (
	"biameet.ir/db"
	"biameet.ir/models"
)

func GetAdminStats() (*models.AdminStats, error) {
	stats := &models.AdminStats{}
	err := db.DB.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM sessions),
			(SELECT COUNT(*) FROM timeslots),
			(SELECT COUNT(*) FROM votes)
	`).Scan(&stats.TotalSessions, &stats.TotalTimeslots, &stats.TotalVotes)
	if err != nil {
		return nil, err
	}
	return stats, nil
}
