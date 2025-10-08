package main

import (
	"alumni-management-system/config"
	"alumni-management-system/repositories"
	"alumni-management-system/routes"
	"alumni-management-system/services"
	"log"
	"os"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to database
	config.ConnectDB()
	defer config.CloseDB()

	// Initialize repositories
	alumniRepo := repositories.NewAlumniRepository()
	pekerjaanRepo := repositories.NewPekerjaanRepository()
	userRepo := repositories.NewUserRepository()

	// Initialize services
	alumniService := services.NewAlumniService(alumniRepo)
	pekerjaanService := services.NewPekerjaanService(pekerjaanRepo, alumniRepo)
	authService := services.NewAuthService(userRepo)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": "Internal Server Error",
				"error":   err.Error(),
			})
		},
	})

	// Global middleware
	app.Use(recover.New()) // Recover from panics
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))

	// Setup routes
	routes.SetupRoutes(app, alumniService, pekerjaanService, authService) // Pass services directly

	// Get port from environment or use default
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "3000"
	}

	// Start server
	log.Printf(" Alumni Management System v2.0 starting on port %s", port)
	log.Printf("Features: CRUD Operations, JWT Authentication, Role-Based Access Control")
	log.Printf(" Default users: admin/123456, user1/123456")
	log.Fatal(app.Listen(":" + port))
}
