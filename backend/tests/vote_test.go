package tests

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"

	"biameet.ir/api"
	"biameet.ir/db"
	"biameet.ir/models"
	"biameet.ir/services"
	"github.com/gofiber/fiber/v2"
	_ "github.com/mattn/go-sqlite3"
)

func setupVoteApp() *fiber.App {
	app := fiber.New()
	testDB := "test_vote.db"
	os.Remove(testDB)

	if err := db.InitDB(testDB); err != nil {
		panic(err)
	}

	apiGroup := app.Group("/api/v1")
	apiGroup.Post("/sessions/:id/vote", api.VoteHandler)

	return app
}

func TestVote(t *testing.T) {
	app := setupVoteApp()
	defer os.Remove("test_vote.db")

	// Seed data
	req := models.CreateSessionRequest{
		Title:       "Vote Test",
		CreatorName: "Tester",
		Timeslots: []models.TimeslotRequest{
			{StartUTC: "2023-01-01T12:00:00Z", EndUTC: "2023-01-01T13:00:00Z"},
		},
	}
	created, err := services.CreateSession(req)
	if err != nil {
		t.Fatalf("Failed to seed session: %v", err)
	}

	// We need to get the timeslot ID.
	// Since CreateSession doesn't return timeslot IDs in response (only session ID),
	// we need to fetch the session to get timeslots.
	session, err := services.GetSession(created.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}
	if len(session.Timeslots) == 0 {
		t.Fatalf("No timeslots found")
	}
	tsID := session.Timeslots[0].ID

	// Test Vote
	voteReq := models.VoteRequest{
		VoterName: "Voter 1",
		Votes: []models.VoteItem{
			{TimeslotID: tsID, Note: "Yes"},
		},
	}

	body, _ := json.Marshal(voteReq)
	httpReq := httptest.NewRequest("POST", "/api/v1/sessions/"+created.ID+"/vote", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(httpReq)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		buf := make([]byte, 1024)
		n, _ := resp.Body.Read(buf)
		t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(buf[:n]))
	}

	// Test Double Vote
	httpReq2 := httptest.NewRequest("POST", "/api/v1/sessions/"+created.ID+"/vote", bytes.NewReader(body))
	httpReq2.Header.Set("Content-Type", "application/json")

	resp2, err := app.Test(httpReq2)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Should fail or return error
	if resp2.StatusCode != 500 { // We returned 500 for error in handler
		t.Errorf("Expected status 500 for double vote, got %d", resp2.StatusCode)
	}
}
