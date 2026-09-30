// Package web serves the frontend: static assets pre-compressed once at
// startup, and index.html rendered as an html/template with per-page SEO tags.
package web

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"io/fs"
	"mime"
	"os"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/gofiber/fiber/v2"
)

// dist is filled by the frontend build (npm run build in frontend/).
//
//go:embed all:dist
var embedded embed.FS

type Options struct {
	// Dir serves the frontend from disk and re-reads it on every request (development).
	// Empty means use the files embedded in the binary.
	Dir     string
	BaseURL string // e.g. https://biameet.ir, no trailing slash
	Version string
}

// Page describes one rendered HTML page.
type Page struct {
	Kind        string // which view the frontend renders: home, session, admin or notfound
	Title       string
	Description string
	Path        string // canonical path; empty for pages that must not be indexed
	NoIndex     bool
	Home        bool
}

type asset struct {
	body, gz, br []byte
	ctype        string
	etag         string
	version      string // short content hash used for ?v= cache busting
}

type bundle struct {
	assets map[string]*asset // keyed by URL path, e.g. "/assets/app.js"
	tmpl   *template.Template
}

type cachedPage struct {
	year int
	page *asset
}

type Site struct {
	opts   Options
	bundle *bundle // nil in dev mode
	home   atomic.Pointer[cachedPage]
}

func New(opts Options) (*Site, error) {
	opts.BaseURL = strings.TrimRight(opts.BaseURL, "/")
	s := &Site{opts: opts}
	if opts.Dir != "" {
		// Fail fast on a bad path, but reload per request.
		_, err := load(os.DirFS(opts.Dir), false)
		return s, err
	}
	sub, err := fs.Sub(embedded, "dist")
	if err != nil {
		return nil, err
	}
	s.bundle, err = load(sub, true)
	return s, err
}

// NewFromFS builds a Site from an arbitrary filesystem (used by tests).
func NewFromFS(fsys fs.FS, opts Options) (*Site, error) {
	b, err := load(fsys, true)
	if err != nil {
		return nil, err
	}
	opts.BaseURL = strings.TrimRight(opts.BaseURL, "/")
	return &Site{opts: opts, bundle: b}, nil
}

func (s *Site) BaseURL() string { return s.opts.BaseURL }

func (s *Site) current() (*bundle, error) {
	if s.bundle != nil {
		return s.bundle, nil
	}
	return load(os.DirFS(s.opts.Dir), false)
}

var compressible = map[string]bool{
	".js": true, ".css": true, ".svg": true, ".json": true, ".txt": true, ".xml": true, ".webmanifest": true,
}

func load(fsys fs.FS, compress bool) (*bundle, error) {
	b := &bundle{assets: map[string]*asset{}}
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || strings.HasPrefix(d.Name(), ".") || p == "index.html" {
			return err
		}
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		ext := path.Ext(p)
		ctype := mime.TypeByExtension(ext)
		if ctype == "" {
			ctype = "application/octet-stream"
		}
		a := newAsset(data, ctype)
		if compress && compressible[ext] {
			a.compress()
		}
		b.assets["/"+p] = a
		return nil
	})
	if err != nil {
		return nil, err
	}

	src, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		return nil, fmt.Errorf("frontend not built (missing index.html): %w", err)
	}
	b.tmpl, err = template.New("index").Funcs(template.FuncMap{
		// asset returns a cache-busting URL, so assets can be cached forever.
		"asset": func(p string) (string, error) {
			a, ok := b.assets["/"+p]
			if !ok {
				return "", fmt.Errorf("unknown asset %q", p)
			}
			return "/" + p + "?v=" + a.version, nil
		},
	}).Parse(string(src))
	return b, err
}

func newAsset(data []byte, ctype string) *asset {
	sum := sha256.Sum256(data)
	h := hex.EncodeToString(sum[:])
	return &asset{body: data, ctype: ctype, etag: `"` + h[:16] + `"`, version: h[:10]}
}

// compress stores brotli and gzip variants when they are actually smaller.
// Done once, at maximum compression, so requests never spend CPU on it.
func (a *asset) compress() {
	var buf bytes.Buffer
	// A 256 KiB window covers every asset; the default 4 MiB one only costs memory.
	bw := brotli.NewWriterOptions(&buf, brotli.WriterOptions{Quality: brotli.BestCompression, LGWin: 18})
	bw.Write(a.body)
	bw.Close()
	if buf.Len() < len(a.body) {
		a.br = bytes.Clone(buf.Bytes())
	}

	buf.Reset()
	gw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	gw.Write(a.body)
	gw.Close()
	if buf.Len() < len(a.body) {
		a.gz = bytes.Clone(buf.Bytes())
	}
}

// ServeAsset writes the static file at the request path. It reports false if there is none.
func (s *Site) ServeAsset(c *fiber.Ctx) (bool, error) {
	b, err := s.current()
	if err != nil {
		return true, err
	}
	a, ok := b.assets[c.Path()]
	if !ok {
		return false, nil
	}
	cache := "public, max-age=86400"
	if v := c.Query("v"); v != "" && v == a.version {
		cache = "public, max-age=31536000, immutable"
	}
	if s.bundle == nil {
		cache = "no-cache"
	}
	return true, send(c, a, cache)
}

func send(c *fiber.Ctx, a *asset, cacheControl string) error {
	c.Set(fiber.HeaderCacheControl, cacheControl)
	c.Set(fiber.HeaderETag, a.etag)
	if a.gz != nil || a.br != nil {
		c.Set(fiber.HeaderVary, fiber.HeaderAcceptEncoding)
	}
	if c.Get(fiber.HeaderIfNoneMatch) == a.etag {
		return c.SendStatus(fiber.StatusNotModified)
	}
	c.Set(fiber.HeaderContentType, a.ctype)

	ae := c.Get(fiber.HeaderAcceptEncoding)
	switch {
	case a.br != nil && strings.Contains(ae, "br"):
		c.Set(fiber.HeaderContentEncoding, "br")
		return c.Send(a.br)
	case a.gz != nil && strings.Contains(ae, "gzip"):
		c.Set(fiber.HeaderContentEncoding, "gzip")
		return c.Send(a.gz)
	}
	return c.Send(a.body)
}

type pageData struct {
	Page
	BaseURL string
	Version string
	Year    int
}

func (s *Site) render(b *bundle, p Page) ([]byte, error) {
	var buf bytes.Buffer
	err := b.tmpl.Execute(&buf, pageData{
		Page:    p,
		BaseURL: s.opts.BaseURL,
		Version: s.opts.Version,
		Year:    time.Now().Year(),
	})
	return buf.Bytes(), err
}

// Render writes an HTML page with the given status code.
func (s *Site) Render(c *fiber.Ctx, status int, p Page) error {
	b, err := s.current()
	if err != nil {
		return err
	}
	if p.NoIndex {
		c.Set("X-Robots-Tag", "noindex, nofollow")
	}

	// The home page is identical for everyone: render and compress it once
	// (re-rendered only when the footer year changes).
	if p.Home && status == fiber.StatusOK && s.bundle != nil {
		year := time.Now().Year()
		cached := s.home.Load()
		if cached == nil || cached.year != year {
			html, err := s.render(b, p)
			if err != nil {
				return err
			}
			a := newAsset(html, fiber.MIMETextHTMLCharsetUTF8)
			a.compress()
			cached = &cachedPage{year: year, page: a}
			s.home.Store(cached)
		}
		return send(c, cached.page, "no-cache")
	}

	html, err := s.render(b, p)
	if err != nil {
		return err
	}
	c.Status(status)
	c.Set(fiber.HeaderCacheControl, "no-cache")
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	c.Set(fiber.HeaderVary, fiber.HeaderAcceptEncoding)
	if strings.Contains(c.Get(fiber.HeaderAcceptEncoding), "gzip") {
		c.Set(fiber.HeaderContentEncoding, "gzip")
		return c.Send(gzipFast(html))
	}
	return c.Send(html)
}

var gzipPool = sync.Pool{New: func() any {
	w, _ := gzip.NewWriterLevel(nil, gzip.BestSpeed)
	return w
}}

func gzipFast(data []byte) []byte {
	var buf bytes.Buffer
	w := gzipPool.Get().(*gzip.Writer)
	w.Reset(&buf)
	w.Write(data)
	w.Close()
	gzipPool.Put(w)
	return buf.Bytes()
}
