package tests

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"biameet.ir/api"
	"biameet.ir/db"
	"biameet.ir/models"
)

func TestCreateAndGetSession(t *testing.T) {
	e := newEnv(t)
	id, slots := createFixed(t, e)

	if len(id) != 5 {
		t.Fatalf("id %q should be 5 chars", id)
	}
	s := getSession(t, e, id)
	if s.Title != "Planning" || s.Type != "fixed" || len(slots) != 2 {
		t.Fatalf("unexpected session: %+v", s)
	}
	// Slots come back chronologically and in canonical ISO form.
	if s.Timeslots[0].StartUTC != "2030-01-01T10:00:00.000Z" {
		t.Fatalf("first slot = %s, want the earliest, normalized", s.Timeslots[0].StartUTC)
	}
	if s.Timeslots[0].Votes == nil {
		t.Fatal("votes should be an empty array, not null")
	}
}

func TestGetUnknownSession(t *testing.T) {
	e := newEnv(t)
	expectError(t, e.do(t, "GET", "/api/v1/sessions/zzzzz", nil), 404, "session_not_found")
}

func TestCreateSessionValidation(t *testing.T) {
	e := newEnv(t)
	slot := []models.TimeslotRequest{{StartUTC: "2030-01-01T10:00:00Z", EndUTC: "2030-01-01T11:00:00Z"}}
	cases := map[string]models.CreateSessionRequest{
		"missing title": {CreatorName: "a", Timeslots: slot},
		"blank title":   {Title: "   ", CreatorName: "a", Timeslots: slot},
		"long title":    {Title: strings.Repeat("x", 201), CreatorName: "a", Timeslots: slot},
		"no slots":      {Title: "t", CreatorName: "a"},
		"bad type":      {Title: "t", CreatorName: "a", Type: "nope", Timeslots: slot},
		"end before start": {Title: "t", CreatorName: "a", Timeslots: []models.TimeslotRequest{
			{StartUTC: "2030-01-01T11:00:00Z", EndUTC: "2030-01-01T10:00:00Z"}}},
		"not a time": {Title: "t", CreatorName: "a", Timeslots: []models.TimeslotRequest{
			{StartUTC: "tomorrow", EndUTC: "2030-01-01T10:00:00Z"}}},
		"dynamic without config": {Title: "t", CreatorName: "a", Type: "dynamic"},
		"weekly bad day": {Title: "t", CreatorName: "a", Type: "weekly",
			DynamicConfig: &models.DynamicConfig{MinTime: "09:00", MaxTime: "10:00", AllowedDays: []int{9}}},
		"weekly inverted hours": {Title: "t", CreatorName: "a", Type: "weekly",
			DynamicConfig: &models.DynamicConfig{MinTime: "12:00", MaxTime: "10:00", AllowedDays: []int{1}}},
		"expiry in past": {Title: "t", CreatorName: "a", Timeslots: slot, ExpiresAtUTC: "2001-01-01T00:00:00Z"},
	}
	for name, req := range cases {
		t.Run(name, func(t *testing.T) {
			expectError(t, e.do(t, "POST", "/api/v1/sessions", req), 400, "invalid_input")
		})
	}
}

func TestExpiryIsStoredAndEnforced(t *testing.T) {
	e := newEnv(t)
	expires := time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339)
	r := e.do(t, "POST", "/api/v1/sessions", models.CreateSessionRequest{
		Title: "Weekly", CreatorName: "Sara", Type: "weekly", ExpiresAtUTC: expires,
		DynamicConfig: &models.DynamicConfig{MinTime: "09:00", MaxTime: "17:00", AllowedDays: []int{1, 3, 3}},
	})
	expectStatus(t, r, 201)
	id := r.JSON(t)["id"].(string)

	s := getSession(t, e, id)
	if s.ExpiresAtUTC == "" || s.Expired {
		t.Fatalf("expiry not stored: %+v", s)
	}
	if len(s.DynamicConfig.AllowedDays) != 2 {
		t.Fatalf("duplicate days should be removed: %v", s.DynamicConfig.AllowedDays)
	}

	// Force it into the past.
	if _, err := db.DB.Exec(`UPDATE sessions SET expires_at_utc = '2001-01-01T00:00:00.000Z' WHERE id = ?`, id); err != nil {
		t.Fatal(err)
	}
	if !getSession(t, e, id).Expired {
		t.Fatal("session should report expired")
	}
	expectError(t, e.do(t, "POST", "/api/v1/sessions/"+id+"/vote",
		models.VoteRequest{VoterName: "Ali"}), 410, "session_expired")
	expectError(t, e.do(t, "POST", "/api/v1/sessions/"+id+"/timeslots",
		models.TimeslotRequest{StartUTC: "2030-01-01T10:00:00Z", EndUTC: "2030-01-01T11:00:00Z", CreatedBy: "Ali"}), 410, "session_expired")
}

func TestUnknownAPIRouteIsJSON404(t *testing.T) {
	e := newEnv(t)
	expectError(t, e.do(t, "GET", "/api/v1/nope", nil), 404, "not_found")
}

func TestBodyLimit(t *testing.T) {
	e := newEnv(t)
	req := httptest.NewRequest("POST", "/api/v1/sessions", strings.NewReader(`{"title":"`+strings.Repeat("x", 100_000)+`"}`))
	req.Header.Set("Content-Type", "application/json")
	// fasthttp rejects the body while reading it; the in-memory test
	// transport surfaces that as an error (a real client gets a 413).
	resp, err := e.app.Test(req, -1)
	if err == nil && resp.StatusCode != 413 {
		t.Fatalf("oversized body accepted with status %d", resp.StatusCode)
	}
	if err != nil && !strings.Contains(err.Error(), "body size exceeds") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAdminStatsRequiresToken(t *testing.T) {
	e := newEnv(t)
	createFixed(t, e)
	expectError(t, e.do(t, "GET", "/api/v1/admin/stats", nil), 401, "unauthorized")
	expectError(t, e.do(t, "GET", "/api/v1/admin/stats", nil, "Authorization", "Bearer wrong"), 401, "unauthorized")

	r := e.do(t, "GET", "/api/v1/admin/stats", nil, "Authorization", "Bearer admintoken")
	expectStatus(t, r, 200)
	stats := r.JSON(t)
	if stats["total_sessions"].(float64) != 1 || stats["total_timeslots"].(float64) != 2 {
		t.Fatalf("stats: %s", r.Body)
	}
}

// Behind a proxy only the entry the proxy appended identifies the client, so a
// client can't dodge the rate limit by sending its own X-Forwarded-For.
func TestRateLimitUsesProxyAppendedIP(t *testing.T) {
	newEnv(t)
	app := api.NewApp(api.Config{ProxyHeader: "X-Forwarded-For", WriteLimit: 2})
	post := func(xff string) int {
		req := httptest.NewRequest("POST", "/api/v1/sessions", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", xff)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		return resp.StatusCode
	}
	post("1.1.1.1, 203.0.113.9")
	post("2.2.2.2, 203.0.113.9")
	if got := post("3.3.3.3, 203.0.113.9"); got != 429 {
		t.Fatalf("spoofed first hop bypassed the limit: status %d", got)
	}
	if got := post("198.51.100.7"); got == 429 {
		t.Fatal("a different client was limited")
	}
}
