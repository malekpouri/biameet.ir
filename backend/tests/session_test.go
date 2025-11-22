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
	"github.com/gofiber/fiber/v2"
	_ "modernc.org/sqlite"
)

func setupApp() *fiber.App {
	app := fiber.New()

	// Init DB with a temp file
	// For simplicity in this environment, we use a file in /tmp
	// or just a local test.db that we clean up
	testDB := "test_biameet.db"
	os.Remove(testDB) // Clean up previous run

	if err := db.InitDB(testDB); err != nil {
		panic(err)
	}

	apiGroup := app.Group("/api/v1")
	apiGroup.Post("/sessions", api.CreateSessionHandler)

	return app
}

func TestCreateSession(t *testing.T) {
	// Ensure we are in the right directory to find migrations if needed
	// But db.InitDB handles relative paths.
	// If running from backend root:
	// db.InitDB will look for db/migrations/...

	app := setupApp()
	defer os.Remove("test_biameet.db")

	payload := models.CreateSessionRequest{
		Title:       "Test Meeting",
		CreatorName: "Tester",
		Timeslots: []models.TimeslotRequest{
			{StartUTC: "2023-01-01T10:00:00Z", EndUTC: "2023-01-01T11:00:00Z"},
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/sessions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 201 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		t.Errorf("Expected status 201, got %d. Body: %s", resp.StatusCode, buf.String())
	}

	var respBody models.CreateSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if respBody.ID == "" {
		t.Error("Expected ID to be returned")
	}
}
