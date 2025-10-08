package middleware

import (
    "alumni-management-system/utils"
    "strings"
	"time"
	"fmt"
    "github.com/gofiber/fiber/v2"
)

// AuthRequired middleware - memverifikasi JWT token
func AuthRequired() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Ambil token dari header Authorization
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return c.Status(401).JSON(fiber.Map{
                "success": false,
                "message": "Token akses diperlukan",
                "error":   "Authorization header missing",
            })
        }

        // Extract token dari "Bearer TOKEN"
        tokenParts := strings.Split(authHeader, " ")
        if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
            return c.Status(401).JSON(fiber.Map{
                "success": false,
                "message": "Format token tidak valid",
                "error":   "Invalid authorization header format",
            })
        }

        tokenString := tokenParts[1]
        if tokenString == "" {
            return c.Status(401).JSON(fiber.Map{
                "success": false,
                "message": "Token tidak boleh kosong",
                "error":   "Empty token",
            })
        }

        // Validasi token
        claims, err := utils.ValidateToken(tokenString)
        if err != nil {
            return c.Status(401).JSON(fiber.Map{
                "success": false,
                "message": "Token tidak valid atau expired",
                "error":   err.Error(),
            })
        }

        // Simpan informasi user di context untuk digunakan di handler
        c.Locals("user_id", claims.UserID)
        c.Locals("username", claims.Username)
        c.Locals("role", claims.Role)
        c.Locals("claims", claims)

        return c.Next()
    }
}

// AdminOnly middleware - memastikan user memiliki role admin
func AdminOnly() fiber.Handler {
    return func(c *fiber.Ctx) error {
        role := c.Locals("role")
        if role == nil {
            return c.Status(401).JSON(fiber.Map{
                "success": false,
                "message": "Informasi user tidak ditemukan",
                "error":   "User context missing",
            })
        }

        userRole, ok := role.(string)
        if !ok || userRole != "admin" {
            return c.Status(403).JSON(fiber.Map{
                "success": false,
                "message": "Akses ditolak. Hanya admin yang diizinkan",
                "error":   "Insufficient privileges",
            })
        }

        return c.Next()
    }
}

// UserOrAdmin middleware - membolehkan user atau admin
func UserOrAdmin() fiber.Handler {
    return func(c *fiber.Ctx) error {
        role := c.Locals("role")
        if role == nil {
            return c.Status(401).JSON(fiber.Map{
                "success": false,
                "message": "Informasi user tidak ditemukan",
                "error":   "User context missing",
            })
        }

        userRole, ok := role.(string)
        if !ok || (userRole != "admin" && userRole != "user") {
            return c.Status(403).JSON(fiber.Map{
                "success": false,
                "message": "Akses ditolak. Role tidak valid",
                "error":   "Invalid role",
            })
        }

        return c.Next()
    }
}

// Optional: Middleware untuk logging akses
func LogAccess() fiber.Handler {
    return func(c *fiber.Ctx) error {
        username := c.Locals("username")
        role := c.Locals("role")
        
        if username != nil && role != nil {
            method := c.Method()
            path := c.Path()
             fmt.Printf("[%s] %s:%s accessed %s %s\n", 
                 time.Now().Format("2006-01-02 15:04:05"), 
                 username, role, method, path)
        }
        
        return c.Next()
    }
}