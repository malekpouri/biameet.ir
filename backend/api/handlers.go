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
	if len(req.Timeslots) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one timeslot is required",
		})
	}

	resp, err := services.CreateSession(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}
