package tests

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"

	"biameet.ir/api"
	"biameet.ir/db"
	"biameet.ir/models"
	"biameet.ir/services"
	"github.com/gofiber/fiber/v2"
	_ "modernc.org/sqlite"
)

func setupGetApp() *fiber.App {
	app := fiber.New()
	// Use the same DB file name to potentially share state if needed,
	// but here we want isolation.
	testDB := "test_get_session.db"
	os.Remove(testDB)

	if err := db.InitDB(testDB); err != nil {
		panic(err)
	}

	apiGroup := app.Group("/api/v1")
	apiGroup.Get("/sessions/:id", api.GetSessionHandler)

	return app
}

func TestGetSession(t *testing.T) {
	app := setupGetApp()
	defer os.Remove("test_get_session.db")

	// Seed data
	req := models.CreateSessionRequest{
		Title:       "Get Test",
		CreatorName: "Tester",
		Timeslots: []models.TimeslotRequest{
			{StartUTC: "2023-01-01T12:00:00Z", EndUTC: "2023-01-01T13:00:00Z"},
		},
	}
	created, err := services.CreateSession(req)
	if err != nil {
		t.Fatalf("Failed to seed session: %v", err)
	}

	// Test Get
	httpReq := httptest.NewRequest("GET", "/api/v1/sessions/"+created.ID, nil)
	resp, err := app.Test(httpReq)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		// Read body
		buf := make([]byte, 1024)
		n, _ := resp.Body.Read(buf)
		t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(buf[:n]))
	}

	var session models.Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if session.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, session.ID)
	}
	if len(session.Timeslots) != 1 {
		t.Errorf("Expected 1 timeslot, got %d", len(session.Timeslots))
	}
}
