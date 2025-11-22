package main

import (
	"log"
	"os"

	"biameet.ir/api"
	"biameet.ir/db"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New()

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Initialize Database
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "biameet.db"
	}
	if err := db.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Routes
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	v1 := app.Group("/api/v1")
	v1.Post("/sessions", api.CreateSessionHandler)
	v1.Get("/sessions/:id", api.GetSessionHandler)
	v1.Post("/sessions/:id/vote", api.VoteHandler)
	v1.Post("/sessions/:id/timeslots", api.AddTimeslotHandler)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Listen(":" + port))
}
