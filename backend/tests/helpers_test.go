package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"testing/fstest"

	"biameet.ir/api"
	"biameet.ir/db"
	"biameet.ir/models"
	"biameet.ir/services"
	"biameet.ir/web"
	"github.com/gofiber/fiber/v2"
)

// A cut-down index.html using the same template fields as frontend/src/index.html.
const testIndex = `<!DOCTYPE html><html><head>
<title>{{.Title}}</title>
<meta name="description" content="{{.Description}}">
{{if .NoIndex}}<meta name="robots" content="noindex, nofollow">{{end}}
{{if .Path}}<link rel="canonical" href="{{.BaseURL}}{{.Path}}">{{end}}
<meta property="og:title" content="{{.Title}}">
<script src="{{asset "assets/app.js"}}" defer></script>
</head><body>{{if .Home}}<section id="landing">landing</section>{{end}}</body></html>`

type testEnv struct {
	app *fiber.App
}

func newEnv(t *testing.T) *testEnv {
	t.Helper()
	if err := db.InitDB(filepath.Join(t.TempDir(), "test.db")); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	services.SetTokenKey([]byte("test-key-0123456789"))

	site, err := web.NewFromFS(fstest.MapFS{
		"index.html":    {Data: []byte(testIndex)},
		"assets/app.js": {Data: bytes.Repeat([]byte("console.log('biameet');\n"), 50)},
	}, web.Options{BaseURL: "https://example.test"})
	if err != nil {
		t.Fatalf("site: %v", err)
	}
	return &testEnv{app: api.NewApp(api.Config{Site: site, AdminToken: "admintoken"})}
}

type response struct {
	Status int
	Header map[string]string
	Body   []byte
}

func (r response) JSON(t *testing.T) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(r.Body, &m); err != nil {
		t.Fatalf("decode %q: %v", r.Body, err)
	}
	return m
}

func (e *testEnv) do(t *testing.T, method, path string, body any, headers ...string) response {
	t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rd)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	resp, err := e.app.Test(req, -1)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	h := map[string]string{}
	for k := range resp.Header {
		h[k] = resp.Header.Get(k)
	}
	return response{Status: resp.StatusCode, Header: h, Body: b}
}

func expectStatus(t *testing.T, r response, want int) {
	t.Helper()
	if r.Status != want {
		t.Fatalf("status = %d, want %d; body: %s", r.Status, want, r.Body)
	}
}

func expectError(t *testing.T, r response, status int, code string) {
	t.Helper()
	expectStatus(t, r, status)
	if got := r.JSON(t)["error"]; got != code {
		t.Fatalf("error code = %v, want %s", got, code)
	}
}

func createFixed(t *testing.T, e *testEnv) (sessionID string, slots []string) {
	t.Helper()
	r := e.do(t, "POST", "/api/v1/sessions", models.CreateSessionRequest{
		Title:       "Planning",
		CreatorName: "Sara",
		Timeslots: []models.TimeslotRequest{
			{StartUTC: "2030-01-02T10:00:00.000Z", EndUTC: "2030-01-02T11:00:00.000Z"},
			{StartUTC: "2030-01-01T10:00:00Z", EndUTC: "2030-01-01T11:00:00Z"},
		},
	})
	expectStatus(t, r, 201)
	sessionID = r.JSON(t)["id"].(string)
	return sessionID, slotIDs(t, e, sessionID)
}

func createDynamic(t *testing.T, e *testEnv) string {
	t.Helper()
	r := e.do(t, "POST", "/api/v1/sessions", models.CreateSessionRequest{
		Title:       "Range",
		CreatorName: "Sara",
		Type:        "dynamic",
		DynamicConfig: &models.DynamicConfig{
			DateUTC: "2030-01-01T00:00:00.000Z", MinTime: "09:00", MaxTime: "17:00",
		},
	})
	expectStatus(t, r, 201)
	return r.JSON(t)["id"].(string)
}

func getSession(t *testing.T, e *testEnv, id string) models.Session {
	t.Helper()
	r := e.do(t, "GET", "/api/v1/sessions/"+id, nil)
	expectStatus(t, r, 200)
	var s models.Session
	if err := json.Unmarshal(r.Body, &s); err != nil {
		t.Fatal(err)
	}
	return s
}

func slotIDs(t *testing.T, e *testEnv, sessionID string) []string {
	var ids []string
	for _, ts := range getSession(t, e, sessionID).Timeslots {
		ids = append(ids, ts.ID)
	}
	return ids
}
