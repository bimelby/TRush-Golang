package routes

import (
	"alumni-management-system/middleware"
	"alumni-management-system/services"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App,
	alumniService services.AlumniService,
	pekerjaanService services.PekerjaanService,
	authService services.AuthService) {

	// API group
	api := app.Group("/alumni-management-system")

	// Health check (public)
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success":  true,
			"message":  "Alumni Management System API is running!",
			"version":  "2.0.0",
			"features": []string{"CRUD", "JWT Auth", "RBAC", "Pagination", "Search", "Sorting"},
		})
	})

	// Authentication routes (public)
	auth := api.Group("/auth")
	auth.Post("/login", authService.Login)
	auth.Post("/register", authService.Register)

	// Protected auth routes (require authentication)
	authProtected := auth.Group("", middleware.AuthRequired())
	authProtected.Get("/profile", authService.GetProfile) // Langsung panggil service method
	authProtected.Post("/logout", authService.Logout)
	authProtected.Get("/validate", func(c *fiber.Ctx) error {
		// Untuk validate, ambil token dari header dan panggil service
		token := c.Get("Authorization")
		claims, err := authService.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Token tidak valid", "error": err.Error()})
		}
		return c.JSON(fiber.Map{"success": true, "message": "Token valid", "data": claims})
	})

	// Protected routes - require authentication
	protected := api.Group("", middleware.AuthRequired())

	// Alumni routes dengan RBAC
	alumni := protected.Group("/alumni")
	// Read operations - Admin dan User bisa akses
	alumni.Get("/", middleware.UserOrAdmin(), alumniService.GetAllAlumni)
	alumni.Get("/without-jobs", middleware.UserOrAdmin(), alumniService.GetAlumniWithoutPekerjaan)
	alumni.Get("/:id", middleware.UserOrAdmin(), alumniService.GetAlumniByID)

	// Write operations - Hanya Admin
	alumni.Post("/", middleware.AdminOnly(), alumniService.CreateAlumni)      // Langsung panggil service method
	alumni.Put("/:id", middleware.AdminOnly(), alumniService.UpdateAlumni)    // Langsung panggil service method
	alumni.Delete("/:id", middleware.AdminOnly(), alumniService.DeleteAlumni) // Langsung panggil service method

	// Pekerjaan routes dengan RBAC
	pekerjaan := protected.Group("/pekerjaan")
	// Read operations - Admin dan User bisa akses
	pekerjaan.Get("/", middleware.UserOrAdmin(), pekerjaanService.GetAllPekerjaan)     
	pekerjaan.Get("/trash", middleware.AdminOnly(), pekerjaanService.GetTrashedPekerjaan)
	
	
	// Special read operation - Hanya Admin
	pekerjaan.Get("/alumni/:alumni_id", middleware.AdminOnly(), pekerjaanService.GetPekerjaanByAlumniID) 

	pekerjaan.Delete("/trash/:id", middleware.AdminOnly(), pekerjaanService.HardDeleteTrashedPekerjaan)  // Hard delete
    pekerjaan.Put("/trash/restore/:id", middleware.AdminOnly(), pekerjaanService.RestoreTrashedPekerjaan)
	
	pekerjaan.Get("/:id", middleware.UserOrAdmin(), pekerjaanService.GetPekerjaanByID) // Langsung panggil service method
	
	// Write operations - Hanya Admin
	pekerjaan.Post("/", middleware.AdminOnly(), pekerjaanService.CreatePekerjaan)      // Langsung panggil service method
	pekerjaan.Put("/:id", middleware.AdminOnly(), pekerjaanService.UpdatePekerjaan)    // Langsung panggil service method
	pekerjaan.Delete("/:id", middleware.AdminOnly(), pekerjaanService.DeletePekerjaan)
	pekerjaan.Delete("/soft-delete/:id", pekerjaanService.SoftDeletePekerjaan) // Langsung panggil service method

}
