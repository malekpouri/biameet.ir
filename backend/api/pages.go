package api

import (
	"errors"
	"regexp"

	"biameet.ir/services"
	"biameet.ir/web"
	"github.com/gofiber/fiber/v2"
)

const (
	siteName        = "بیا میت"
	homeTitle       = "بیا میت | هماهنگی و زمان‌بندی رایگان جلسات"
	homeDescription = "بیا میت ابزاری ساده، رایگان و متن‌باز برای هماهنگی زمان جلسات است. بدون ثبت‌نام چند زمان پیشنهاد دهید، لینک را بفرستید و ببینید چه زمانی برای همه مناسب است."
)

// Must match the frontend's getSessionIDFromURL() and utils.GenerateShortID(5).
var sessionPath = regexp.MustCompile(`^/[a-zA-Z0-9]{5}$`)

// PagesHandler serves static assets and every HTML page.
func PagesHandler(site *web.Site) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if ok, err := site.ServeAsset(c); ok || err != nil {
			return err
		}

		p := c.Path()
		switch {
		case p == "/":
			return site.Render(c, fiber.StatusOK, web.Page{
				Kind: "home", Title: homeTitle, Description: homeDescription, Path: "/", Home: true,
			})

		case p == "/admin":
			return site.Render(c, fiber.StatusOK, web.Page{
				Kind: "admin", Title: "داشبورد مدیریت | " + siteName, Description: homeDescription, NoIndex: true,
			})

		case sessionPath.MatchString(p):
			// Private invitation: rich link previews, but never indexed.
			title, creator, err := services.GetSessionSummary(p[1:])
			if errors.Is(err, services.ErrSessionNotFound) {
				return notFound(c, site)
			}
			if err != nil {
				return err
			}
			return site.Render(c, fiber.StatusOK, web.Page{
				Kind:        "session",
				Title:       title + " | " + siteName,
				Description: "دعوت به جلسه توسط " + creator + " — زمان‌های مناسب خود را انتخاب کنید.",
				NoIndex:     true,
			})
		}
		return notFound(c, site)
	}
}

func notFound(c *fiber.Ctx, site *web.Site) error {
	return site.Render(c, fiber.StatusNotFound, web.Page{
		Kind: "notfound", Title: "صفحه یافت نشد | " + siteName, Description: homeDescription, NoIndex: true,
	})
}

func RobotsHandler(site *web.Site) fiber.Handler {
	body := "User-agent: *\nDisallow: /api/\nDisallow: /admin\n\nSitemap: " + site.BaseURL() + "/sitemap.xml\n"
	return func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderCacheControl, "public, max-age=86400")
		c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
		return c.SendString(body)
	}
}

func SitemapHandler(site *web.Site) fiber.Handler {
	body := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>` + site.BaseURL() + `/</loc><changefreq>monthly</changefreq><priority>1.0</priority></url>
</urlset>
`
	return func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderCacheControl, "public, max-age=86400")
		c.Set(fiber.HeaderContentType, "application/xml; charset=utf-8")
		return c.SendString(body)
	}
}
