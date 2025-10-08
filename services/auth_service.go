package services

import (
	"alumni-management-system/models"
	"alumni-management-system/repositories"
	"alumni-management-system/utils"
	"errors"
	"fmt"
	"strings"
	

	"github.com/gofiber/fiber/v2"
)

type AuthService interface {
    Login(c *fiber.Ctx) error
    Register(c *fiber.Ctx) error
    GetProfile(c *fiber.Ctx) error // Updated: now takes *fiber.Ctx and returns error
    Logout(c *fiber.Ctx) error
    ValidateToken(tokenString string) (*models.JWTClaims, error) // Keep as is, used in middleware
}

type authService struct {
    userRepo repositories.UserRepository
}

func NewAuthService(userRepo repositories.UserRepository) AuthService {
    return &authService{
        userRepo: userRepo,
    }
}

// Login - authenticate user dan generate token
func (s *authService) Login(c *fiber.Ctx) error {
    var req models.LoginRequest

    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "Request body tidak valid",
            "error":   err.Error(),
        })
    }

    fmt.Printf("[DEBUG] Login attempt - Username: %s\n", req.Username)

    // Cari user berdasarkan username atau email
    var user *models.User
    var passwordHash string
    var err error

    // Check if input is email or username
    if strings.Contains(req.Username, "@") {
        user, passwordHash, err = s.userRepo.GetByEmail(req.Username)
    } else {
        user, passwordHash, err = s.userRepo.GetByUsername(req.Username)
    }

    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{ // Return 500 for internal server error
            "success": false,
            "message": "Error saat mencari user",
            "error":   err.Error(),
        })
    }

    if user == nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "success": false,
            "message": "Username atau password salah",
        })
    }

    fmt.Printf("[DEBUG] User found - ID: %d, Username: %s, Role: %s\n", user.ID, user.Username, user.Role)
    fmt.Printf("[DEBUG] Password hash from DB: %s\n", passwordHash)

    // Verify password
    if !utils.CheckPassword(req.Password, passwordHash) {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "success": false,
            "message": "Username atau password salah",
        })
    }

    // Generate JWT token
    token, err := utils.GenerateToken(*user)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": "Gagal generate token",
            "error":   err.Error(),
        })
    }

    // Update last login (optional)
    s.userRepo.UpdateLastLogin(user.ID)

    response := &models.LoginResponse{
        User:  *user,
        Token: token,
    }

    fmt.Printf("[DEBUG] Login successful - User: %s, Role: %s\n", response.User.Username, response.User.Role)

    return c.JSON(fiber.Map{
        "success": true,
        "message": "Login berhasil",
        "data":    response,
    })
}

// Register - registrasi user baru (optional)
func (s *authService) Register(c *fiber.Ctx) error {
    var req models.RegisterRequest

    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "Request body tidak valid",
            "error":   err.Error(),
        })
    }

    // Validasi input
    if req.Username == "" || req.Email == "" || req.Password == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "Semua field harus diisi",
        })
    }

    if !utils.ValidatePasswordStrength(req.Password) {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "Password minimal 6 karakter",
        })
    }

    // Set default role jika tidak ada
    if req.Role == "" {
        req.Role = "user"
    }

    // Validasi role
    if req.Role != "admin" && req.Role != "user" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "Role harus admin atau user",
        })
    }

    // Check if username already exists
    existingUser , _, err := s.userRepo.GetByUsername(req.Username)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": "Error saat check username",
            "error":   err.Error(),
        })
    }
    if existingUser  != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "Username sudah digunakan",
        })
    }

    // Check if email already exists
    existingUser , _, err = s.userRepo.GetByEmail(req.Email)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": "Error saat check email",
            "error":   err.Error(),
        })
    }
    if existingUser  != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "Email sudah digunakan",
        })
    }

    // Hash password
    passwordHash, err := utils.HashPassword(req.Password)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": "Gagal hash password",
            "error":   err.Error(),
        })
    }

    // Create user
    user, err := s.userRepo.Create(&req, passwordHash)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": "Gagal membuat user",
            "error":   err.Error(),
        })
    }

    return c.Status(fiber.StatusCreated).JSON(fiber.Map{
        "success": true,
        "message": "Registrasi berhasil",
        "data":    user,
    })
}

// GetProfile - ambil profile user berdasarkan ID dari context
func (s *authService) GetProfile(c *fiber.Ctx) error {
    userID, ok := c.Locals("user_id").(int)
    if !ok {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "success": false,
            "message": "User ID tidak ditemukan di context",
        })
    }

    user, err := s.userRepo.GetByID(userID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": "Error saat mengambil profile",
            "error":   err.Error(),
        })
    }

    if user == nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "success": false,
            "message": "User tidak ditemukan",
        })
    }

    profile := &models.ProfileResponse{
        UserID:   user.ID,
        Username: user.Username,
        Role:     user.Role,
        Email:    user.Email,
    }

    return c.JSON(fiber.Map{
        "success": true,
        "message": "Profile berhasil diambil",
        "data":    profile,
    })
}

// Logout - (optional - client-side token removal)
func (s *authService) Logout(c *fiber.Ctx) error {
    // Dalam JWT, logout biasanya dilakukan di client-side dengan menghapus token
    // Server-side logout memerlukan token blacklist yang lebih kompleks
    
    return c.JSON(fiber.Map{
        "success": true,
        "message": "Logout berhasil. Hapus token dari client",
    })
}

// ValidateToken - validasi JWT token (digunakan di middleware, bukan di route handler)
func (s *authService) ValidateToken(tokenString string) (*models.JWTClaims, error) {
    // Extract token from "Bearer TOKEN" if present
    tokenParts := strings.Split(tokenString, " ")
    if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
        tokenString = tokenParts[1]
    } else if len(tokenParts) != 1 {
        return nil, errors.New("invalid authorization header format")
    }

    claims, err := utils.ValidateToken(tokenString)
    if err != nil {
        return nil, err
    }

    // Optional: Check if token expired (sudah di-handle di utils.ValidateToken)
    // if time.Now().Unix() > claims.Exp { // claims.Exp is int64, time.Now().Unix() is int64
    //     return nil, errors.New("token expired")
    // }

    return claims, nil
}