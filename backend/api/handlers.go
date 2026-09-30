package api

import (
	"crypto/subtle"
	"errors"
	"log"
	"strings"

	"biameet.ir/models"
	"biameet.ir/services"
	"github.com/gofiber/fiber/v2"
)

// fail writes a service error as JSON: {"error": code, "message": text}.
// Unknown errors are logged and hidden behind a generic 500.
func fail(c *fiber.Ctx, err error) error {
	var se *services.Error
	if errors.As(err, &se) {
		return c.Status(se.Status).JSON(fiber.Map{"error": se.Code, "message": se.Message})
	}
	var fe *fiber.Error
	if errors.As(err, &fe) {
		return c.Status(fe.Code).JSON(fiber.Map{"error": "request_error", "message": fe.Message})
	}
	log.Printf("ERROR %s %s: %v", c.Method(), c.Path(), err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error":   "internal_error",
		"message": "خطای داخلی سرور، لطفاً دوباره تلاش کنید",
	})
}

func badBody(c *fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid_body", "message": "درخواست نامعتبر است"})
}

func CreateSessionHandler(c *fiber.Ctx) error {
	var req models.CreateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return badBody(c)
	}
	resp, err := services.CreateSession(req)
	if err != nil {
		return fail(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func GetSessionHandler(c *fiber.Ctx) error {
	session, err := services.GetSession(c.Params("id"))
	if err != nil {
		return fail(c, err)
	}
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.JSON(session)
}

func AddTimeslotHandler(c *fiber.Ctx) error {
	var req models.TimeslotRequest
	if err := c.BodyParser(&req); err != nil {
		return badBody(c)
	}
	ts, err := services.AddTimeslot(c.Params("id"), req)
	if err != nil {
		return fail(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(ts)
}

func DeleteTimeslotHandler(c *fiber.Ctx) error {
	var req models.DeleteTimeslotRequest
	// An empty body is fine (unprotected slot).
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return badBody(c)
		}
	}
	if err := services.DeleteTimeslot(c.Params("id"), c.Params("ts_id"), req); err != nil {
		return fail(c, err)
	}
	return c.JSON(fiber.Map{"status": "ok"})
}

func VoteHandler(c *fiber.Ctx) error {
	var req models.VoteRequest
	if err := c.BodyParser(&req); err != nil {
		return badBody(c)
	}
	token, err := services.SubmitVote(c.Params("id"), req)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(models.VoteResponse{Status: "ok", Token: token})
}

// AdminStatsHandler requires "Authorization: Bearer <ADMIN_TOKEN>".
// With no token configured the endpoint doesn't exist.
func AdminStatsHandler(adminToken string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if adminToken == "" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not_found", "message": "یافت نشد"})
		}
		got := strings.TrimPrefix(c.Get(fiber.HeaderAuthorization), "Bearer ")
		if subtle.ConstantTimeCompare([]byte(got), []byte(adminToken)) != 1 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized", "message": "توکن مدیریت نامعتبر است"})
		}
		stats, err := services.GetAdminStats()
		if err != nil {
			return fail(c, err)
		}
		c.Set(fiber.HeaderCacheControl, "no-store")
		return c.JSON(stats)
	}
}
