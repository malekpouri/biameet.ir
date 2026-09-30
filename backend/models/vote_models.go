package models

type VoteRequest struct {
	VoterName string     `json:"voter_name"`
	Password  string     `json:"password,omitempty"`
	Token     string     `json:"token,omitempty"`
	Votes     []VoteItem `json:"votes"`
}

type VoteItem struct {
	TimeslotID string `json:"timeslot_id"`
	Note       string `json:"note,omitempty"`
}

type VoteResponse struct {
	Status string `json:"status"`
	// Token lets the client edit this vote later without storing the password.
	Token string `json:"token,omitempty"`
}
