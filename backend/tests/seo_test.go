package tests

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"

	"biameet.ir/api"
	"biameet.ir/web"
)

// newRealSiteApp renders the real frontend/src/index.html (with placeholder
// assets), so the SEO markup Google sees is what gets tested.
func newRealSiteApp(t *testing.T, verification string) *testEnv {
	t.Helper()
	newEnv(t) // database
	index, err := os.ReadFile("../../frontend/src/index.html")
	if err != nil {
		t.Fatal(err)
	}
	site, err := web.NewFromFS(fstest.MapFS{
		"index.html":             {Data: index},
		"assets/app.js":          {Data: []byte("//")},
		"assets/app.css":         {Data: []byte("/**/")},
		"assets/vazirmatn.woff2": {Data: []byte("wOF2")},
		"favicon.ico":            {Data: []byte{0, 0, 1, 0}},
		"site.webmanifest":       {Data: []byte(`{"name":"x"}`)},
	}, web.Options{BaseURL: "https://www.example.test", Version: "9.9.9", GoogleVerification: verification})
	if err != nil {
		t.Fatalf("real index.html failed to parse: %v", err)
	}
	return &testEnv{app: api.NewApp(api.Config{Site: site})}
}

var ldJSON = regexp.MustCompile(`(?s)<script type="application/ld\+json">(.*?)</script>`)

func TestHomeStructuredDataIsValidJSON(t *testing.T) {
	e := newRealSiteApp(t, "")
	html := string(e.do(t, "GET", "/", nil).Body)

	m := ldJSON.FindStringSubmatch(html)
	if m == nil {
		t.Fatal("home page has no JSON-LD")
	}
	var doc struct {
		Graph []map[string]any `json:"@graph"`
	}
	if err := json.Unmarshal([]byte(m[1]), &doc); err != nil {
		t.Fatalf("JSON-LD does not parse (Google would ignore it): %v\n%s", err, m[1])
	}
	types := map[string]map[string]any{}
	for _, node := range doc.Graph {
		types[node["@type"].(string)] = node
	}
	for _, want := range []string{"WebSite", "Organization", "WebApplication", "FAQPage"} {
		if types[want] == nil {
			t.Errorf("JSON-LD missing %s", want)
		}
	}
	if got := types["WebSite"]["url"]; got != "https://www.example.test/" {
		t.Errorf("WebSite url = %v", got)
	}
	if got := types["Organization"]["logo"]; got != "https://www.example.test/icon-512.png" {
		t.Errorf("Organization logo = %v", got)
	}
	if n := len(types["FAQPage"]["mainEntity"].([]any)); n != 5 {
		t.Errorf("FAQ has %d questions, want 5 (must match the visible FAQ)", n)
	}
}

func TestHomeSEOTags(t *testing.T) {
	e := newRealSiteApp(t, "abc123-token")
	html := string(e.do(t, "GET", "/", nil).Body)
	for _, want := range []string{
		`<link rel="canonical" href="https://www.example.test/">`,
		`<meta property="og:url" content="https://www.example.test/">`,
		`<meta name="google-site-verification" content="abc123-token">`,
		`<link rel="icon" href="/favicon.ico" sizes="48x48">`,
		`<link rel="manifest" href="/site.webmanifest">`,
		`<html lang="fa" dir="rtl">`,
		`v9.9.9`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("home page missing %s", want)
		}
	}
	if strings.Contains(html, `name="robots"`) {
		t.Error("home page must not carry a robots meta tag")
	}

	// The verification tag belongs on the home page only.
	if strings.Contains(string(e.do(t, "GET", "/zzzzz", nil).Body), "google-site-verification") {
		t.Error("verification tag leaked onto other pages")
	}
	// Without a token there is no empty tag.
	plain := newRealSiteApp(t, "")
	if strings.Contains(string(plain.do(t, "GET", "/", nil).Body), "google-site-verification") {
		t.Error("empty verification tag rendered")
	}
}

func TestAssetContentTypes(t *testing.T) {
	e := newRealSiteApp(t, "")
	for path, want := range map[string]string{
		"/assets/vazirmatn.woff2": "font/woff2",
		"/favicon.ico":            "image/x-icon",
		"/site.webmanifest":       "application/manifest+json",
		"/assets/app.css":         "text/css; charset=utf-8",
	} {
		if got := e.do(t, "GET", path, nil).Header["Content-Type"]; got != want {
			t.Errorf("%s: Content-Type %q, want %q", path, got, want)
		}
	}
}

func TestSitemapAndRobotsUseBaseURL(t *testing.T) {
	e := newRealSiteApp(t, "")
	sm := e.do(t, "GET", "/sitemap.xml", nil)
	expectStatus(t, sm, 200)
	body := string(sm.Body)
	if !strings.Contains(body, "<loc>https://www.example.test/</loc>") || !regexp.MustCompile(`<lastmod>\d{4}-\d{2}-\d{2}</lastmod>`).MatchString(body) {
		t.Fatalf("sitemap: %s", body)
	}
	robots := string(e.do(t, "GET", "/robots.txt", nil).Body)
	if !strings.Contains(robots, "Sitemap: https://www.example.test/sitemap.xml") || !strings.Contains(robots, "Disallow: /api/") {
		t.Fatalf("robots.txt: %s", robots)
	}
	// HEAD must work too (crawlers and uptime checks use it).
	expectStatus(t, e.do(t, "HEAD", "/sitemap.xml", nil), 200)
	expectStatus(t, e.do(t, "HEAD", "/", nil), 200)
}
