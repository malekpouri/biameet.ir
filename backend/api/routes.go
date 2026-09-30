package api

import (
	"errors"
	"log"
	"strings"
	"time"

	"biameet.ir/version"
	"biameet.ir/web"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Config struct {
	Site       *web.Site
	AdminToken string
	// ProxyHeader names the header a reverse proxy uses to pass the client IP
	// (e.g. X-Forwarded-For or X-Real-IP). Only the last entry is trusted: it's
	// the one the proxy itself added; anything before it came from the client.
	ProxyHeader string
	LogRequests bool
	WriteLimit  int // write requests per minute per client IP; 0 disables the limit
}

const securityCSP = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; " +
	"img-src 'self' data:; font-src 'self'; connect-src 'self'; " +
	"base-uri 'none'; form-action 'self'; frame-ancestors 'none'"

func NewApp(cfg Config) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		BodyLimit:             64 * 1024,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          15 * time.Second,
		IdleTimeout:           60 * time.Second,
		// Trades a little CPU for noticeably lower idle memory in fasthttp.
		ReduceMemoryUsage: true,
		ErrorHandler:      errorHandler,
	})

	app.Use(recover.New())
	if cfg.LogRequests {
		app.Use(logger.New())
	}
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Content-Security-Policy", securityCSP)
		return c.Next()
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "version": version.Version})
	})

	v1 := app.Group("/api/v1", compress.New(compress.Config{Level: compress.LevelBestSpeed}))

	write := func(c *fiber.Ctx) error { return c.Next() }
	if cfg.WriteLimit > 0 {
		write = limiter.New(limiter.Config{
			Max:          cfg.WriteLimit,
			Expiration:   time.Minute,
			KeyGenerator: clientIP(cfg.ProxyHeader),
			LimitReached: func(c *fiber.Ctx) error {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error": "rate_limited", "message": "تعداد درخواست‌ها زیاد است، کمی بعد دوباره تلاش کنید",
				})
			},
		})
	}

	v1.Post("/sessions", write, CreateSessionHandler)
	v1.Get("/sessions/:id", GetSessionHandler)
	v1.Post("/sessions/:id/vote", write, VoteHandler)
	v1.Post("/sessions/:id/timeslots", write, AddTimeslotHandler)
	v1.Delete("/sessions/:id/timeslots/:ts_id", write, DeleteTimeslotHandler)
	v1.Get("/admin/stats", write, AdminStatsHandler(cfg.AdminToken))
	v1.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not_found", "message": "یافت نشد"})
	})

	if cfg.Site != nil {
		app.Get("/robots.txt", RobotsHandler(cfg.Site))
		app.Get("/sitemap.xml", SitemapHandler(cfg.Site))
		app.Get("/*", PagesHandler(cfg.Site))
	}
	return app
}

func clientIP(header string) func(*fiber.Ctx) string {
	return func(c *fiber.Ctx) string {
		if header != "" {
			v := c.Get(header)
			if i := strings.LastIndexByte(v, ','); i >= 0 {
				v = v[i+1:]
			}
			if v = strings.TrimSpace(v); v != "" {
				return v
			}
		}
		return c.IP()
	}
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	var fe *fiber.Error
	if errors.As(err, &fe) {
		code = fe.Code
	}
	if strings.HasPrefix(c.Path(), "/api/") {
		return fail(c, err)
	}
	if code == fiber.StatusInternalServerError {
		// Log real failures; don't leak details.
		log.Printf("ERROR %s %s: %v", c.Method(), c.Path(), err)
		return c.Status(code).SendString("خطای داخلی سرور")
	}
	return c.Status(code).SendString(fe.Message)
}
