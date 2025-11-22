package api

import (
	"biameet.ir/models"
	"biameet.ir/services"
	"github.com/gofiber/fiber/v2"
)

func CreateSessionHandler(c *fiber.Ctx) error {
	var req models.CreateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Basic Validation
	if req.Title == "" || req.CreatorName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Title and Creator Name are required",
		})
	}

	// Validation for fixed type
	if req.Type != "dynamic" && len(req.Timeslots) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one timeslot is required for fixed sessions",
		})
	}
	// Validation for dynamic type
	if req.Type == "dynamic" {
		if req.DynamicConfig == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Dynamic config is required for dynamic sessions",
			})
		}
		if req.DynamicConfig.DateUTC == "" || req.DynamicConfig.MinTime == "" || req.DynamicConfig.MaxTime == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Date, MinTime, and MaxTime are required for dynamic sessions",
			})
		}
	}

	resp, err := services.CreateSession(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

func AddTimeslotHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Session ID is required",
		})
	}

	var req models.TimeslotRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.StartUTC == "" || req.EndUTC == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Start and End times are required",
		})
	}

	// TODO: Validate if timeslot is within dynamic session limits (can be done here or service)
	// For now, we trust the frontend or add simple validation later.

	ts, err := services.AddTimeslot(id, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(ts)
}

func GetSessionHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Session ID is required",
		})
	}

	session, err := services.GetSession(id)
	if err != nil {
		if err.Error() == "session not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Session not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(session)
}

func VoteHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Session ID is required",
		})
	}

	var req models.VoteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.VoterName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Voter name is required",
		})
	}

	if len(req.Votes) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one vote is required",
		})
	}

	err := services.SubmitVote(id, req)
	if err != nil {
		if err.Error() == "session not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Session not found",
			})
		}
		// Check for double vote error message pattern if needed, or just 409
		// Simple check
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(models.VoteResponse{Status: "ok"})
}

func GetAdminStatsHandler(c *fiber.Ctx) error {
	stats, err := services.GetAdminStats()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(stats)
}
