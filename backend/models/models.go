package models

type Session struct {
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	CreatorName   string         `json:"creator_name"`
	CreatedAtUTC  string         `json:"created_at_utc"`
	ExpiresAtUTC  string         `json:"expires_at_utc,omitempty"`
	ArchivedAtUTC string         `json:"archived_at_utc,omitempty"`
	Expired       bool           `json:"expired,omitempty"`
	Timeslots     []Timeslot     `json:"timeslots"`
	Type          string         `json:"type"` // "fixed", "dynamic" or "weekly"
	DynamicConfig *DynamicConfig `json:"dynamic_config,omitempty"`
}

type DynamicConfig struct {
	DateUTC     string `json:"date_utc,omitempty"`     // dynamic only: the chosen day at 00:00 UTC
	MinTime     string `json:"min_time"`               // "HH:MM", creator's local time
	MaxTime     string `json:"max_time"`               // "HH:MM", creator's local time
	AllowedDays []int  `json:"allowed_days,omitempty"` // weekly only: JS weekday numbers, 0=Sunday … 6=Saturday
}

type Timeslot struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	StartUTC  string `json:"start_utc"`
	EndUTC    string `json:"end_utc"`
	Votes     []Vote `json:"votes"`
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
	ExpiresAtUTC  string            `json:"expires_at_utc,omitempty"`
}

type TimeslotRequest struct {
	StartUTC  string `json:"start_utc"`
	EndUTC    string `json:"end_utc"`
	CreatedBy string `json:"created_by,omitempty"`
	Password  string `json:"password,omitempty"`
	Token     string `json:"token,omitempty"`
}

type DeleteTimeslotRequest struct {
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
}

type CreateSessionResponse struct {
	ID   string `json:"id"`
	Link string `json:"link"`
}

type AddTimeslotResponse struct {
	Timeslot
	Token string `json:"token,omitempty"`
}

type AdminStats struct {
	TotalSessions  int `json:"total_sessions"`
	TotalTimeslots int `json:"total_timeslots"`
	TotalVotes     int `json:"total_votes"`
}
