package tests

import (
	"strings"
	"testing"

	"biameet.ir/models"
)

func TestSessionPageEscapesUserInput(t *testing.T) {
	e := newEnv(t)
	payload := `"><script>alert(1)</script>`
	r := e.do(t, "POST", "/api/v1/sessions", models.CreateSessionRequest{
		Title: payload, CreatorName: payload,
		Timeslots: []models.TimeslotRequest{{StartUTC: s10, EndUTC: s11}},
	})
	expectStatus(t, r, 201)
	id := r.JSON(t)["id"].(string)

	page := e.do(t, "GET", "/"+id, nil)
	expectStatus(t, page, 200)
	html := string(page.Body)
	if strings.Contains(html, "<script>alert(1)") {
		t.Fatalf("user input was not escaped:\n%s", html)
	}
	if !strings.Contains(html, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Fatalf("escaped title missing:\n%s", html)
	}
	if page.Header["X-Robots-Tag"] != "noindex, nofollow" || !strings.Contains(html, `content="noindex, nofollow"`) {
		t.Fatal("session pages must be noindex")
	}
	if strings.Contains(html, `rel="canonical"`) {
		t.Fatal("private session pages should not declare a canonical URL")
	}
}

func TestHomePage(t *testing.T) {
	e := newEnv(t)
	r := e.do(t, "GET", "/", nil, "Accept-Encoding", "br, gzip")
	expectStatus(t, r, 200)
	if r.Header["Content-Encoding"] != "br" {
		t.Fatalf("home should be served pre-compressed, got %q", r.Header["Content-Encoding"])
	}
	if r.Header["X-Robots-Tag"] != "" {
		t.Fatal("home must be indexable")
	}
	if r.Header["Content-Security-Policy"] == "" {
		t.Fatal("missing CSP header")
	}

	plain := e.do(t, "GET", "/", nil)
	html := string(plain.Body)
	for _, want := range []string{`rel="canonical" href="https://example.test/"`, `id="landing"`, `/assets/app.js?v=`} {
		if !strings.Contains(html, want) {
			t.Errorf("home page missing %q", want)
		}
	}

	etag := plain.Header["Etag"]
	expectStatus(t, e.do(t, "GET", "/", nil, "If-None-Match", etag), 304)
}

func TestNotFoundPages(t *testing.T) {
	e := newEnv(t)
	for _, p := range []string{"/zzzzz", "/does/not/exist", "/banner.jpg"} {
		r := e.do(t, "GET", p, nil)
		expectStatus(t, r, 404)
		if r.Header["X-Robots-Tag"] == "" {
			t.Errorf("%s: 404 pages must be noindex", p)
		}
	}
}

func TestStaticAssets(t *testing.T) {
	e := newEnv(t)
	home := string(e.do(t, "GET", "/", nil).Body)
	i := strings.Index(home, "/assets/app.js?v=")
	url := home[i : i+len("/assets/app.js?v=")+10]

	r := e.do(t, "GET", url, nil, "Accept-Encoding", "gzip")
	expectStatus(t, r, 200)
	if r.Header["Content-Encoding"] != "gzip" || !strings.Contains(r.Header["Cache-Control"], "immutable") {
		t.Fatalf("versioned asset headers: %v", r.Header)
	}
	if !strings.HasPrefix(r.Header["Content-Type"], "text/javascript") && !strings.HasPrefix(r.Header["Content-Type"], "application/javascript") {
		t.Fatalf("content type %q", r.Header["Content-Type"])
	}
	// A stale version must not be cached forever.
	stale := e.do(t, "GET", "/assets/app.js?v=old", nil)
	if strings.Contains(stale.Header["Cache-Control"], "immutable") {
		t.Fatal("stale version marked immutable")
	}
}

func TestRobotsAndSitemap(t *testing.T) {
	e := newEnv(t)
	robots := e.do(t, "GET", "/robots.txt", nil)
	expectStatus(t, robots, 200)
	if !strings.Contains(string(robots.Body), "Sitemap: https://example.test/sitemap.xml") {
		t.Fatalf("robots.txt: %s", robots.Body)
	}
	sitemap := e.do(t, "GET", "/sitemap.xml", nil)
	expectStatus(t, sitemap, 200)
	if !strings.Contains(string(sitemap.Body), "<loc>https://example.test/</loc>") {
		t.Fatalf("sitemap.xml: %s", sitemap.Body)
	}
}
