package main

import (
	"log"

	"biameet.ir/backend/db"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Initialize Database
	db.InitDB()
	defer db.DB.Close()

	// Initialize Fiber
	app := fiber.New()

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Routes
	api := app.Group("/api/v1")
	
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"message": "BiaMeet API is running",
		})
	})

	// Start Server
	log.Fatal(app.Listen(":3000"))
}
