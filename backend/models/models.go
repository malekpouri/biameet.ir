package models

type Session struct {
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	CreatorName   string         `json:"creator_name"`
	CreatedAtUTC  string         `json:"created_at_utc"`
	ExpiresAtUTC  string         `json:"expires_at_utc,omitempty"`
	ArchivedAtUTC string         `json:"archived_at_utc,omitempty"`
	Timeslots     []Timeslot     `json:"timeslots,omitempty"`
	Type          string         `json:"type"` // "fixed" or "dynamic"
	DynamicConfig *DynamicConfig `json:"dynamic_config,omitempty"`
}

type DynamicConfig struct {
	DateUTC     string `json:"date_utc,omitempty"`     // Optional for weekly
	MinTime     string `json:"min_time"`               // "HH:MM"
	MaxTime     string `json:"max_time"`               // "HH:MM"
	AllowedDays []int  `json:"allowed_days,omitempty"` // 0-6 (Sat-Fri or Sun-Sat? Let's assume 0=Saturday as per frontend or just standard 0=Sunday)
	// Frontend used (idx + 6) % 7 for Jalali.
	// Let's store standard JS Day (0=Sunday, 6=Saturday) to be safe, or just what frontend sends.
}

type Timeslot struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	StartUTC  string `json:"start_utc"`
	EndUTC    string `json:"end_utc"`
	Votes     []Vote `json:"votes,omitempty"` // Added Votes
	CreatedBy string `json:"created_by,omitempty"`
}

type Vote struct {
	ID           string `json:"id"`
	TimeslotID   string `json:"timeslot_id"`
	VoterName    string `json:"voter_name"`
	Note         string `json:"note,omitempty"`
	CreatedAtUTC string `json:"created_at_utc"`
}

type CreateSessionRequest struct {
	Title         string            `json:"title"`
	CreatorName   string            `json:"creator_name"`
	Timeslots     []TimeslotRequest `json:"timeslots"`
	Type          string            `json:"type"`
	DynamicConfig *DynamicConfig    `json:"dynamic_config,omitempty"`
}

type TimeslotRequest struct {
	StartUTC  string `json:"start_utc"`
	EndUTC    string `json:"end_utc"`
	CreatedBy string `json:"created_by,omitempty"`
	Password  string `json:"password,omitempty"`
}

type CreateSessionResponse struct {
	ID   string `json:"id"`
	Link string `json:"link"`
}

type AdminStats struct {
	TotalSessions  int `json:"total_sessions"`
	TotalTimeslots int `json:"total_timeslots"`
	TotalVotes     int `json:"total_votes"`
}
