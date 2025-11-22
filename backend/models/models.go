package models

type Session struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	CreatorName   string     `json:"creator_name"`
	CreatedAtUTC  string     `json:"created_at_utc"`
	ExpiresAtUTC  string     `json:"expires_at_utc,omitempty"`
	ArchivedAtUTC string     `json:"archived_at_utc,omitempty"`
	Timeslots     []Timeslot `json:"timeslots,omitempty"`
}

type Timeslot struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	StartUTC  string `json:"start_utc"`
	EndUTC    string `json:"end_utc"`
}

type CreateSessionRequest struct {
	Title       string            `json:"title"`
	CreatorName string            `json:"creator_name"`
	Timeslots   []TimeslotRequest `json:"timeslots"`
}

type TimeslotRequest struct {
	StartUTC string `json:"start_utc"`
	EndUTC   string `json:"end_utc"`
}

type CreateSessionResponse struct {
	ID   string `json:"id"`
	Link string `json:"link"`
}
