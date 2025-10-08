package models

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

// User entity
type User struct {
    ID        int       `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    Role      string    `json:"role"`
    IsDeleted bool      `json:"is_deleted"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Login request DTO
type LoginRequest struct {
    Username string `json:"username" validate:"required"`
    Password string `json:"password" validate:"required"`
}

// Register request DTO (optional untuk future use)
type RegisterRequest struct {
    Username string `json:"username" validate:"required"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
    Role     string `json:"role" validate:"oneof=admin user"`
}

// Login response DTO
type LoginResponse struct {
    User  User   `json:"user"`
    Token string `json:"token"`
}

// JWT Claims structure
type JWTClaims struct {
    UserID   int    `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

// Profile response DTO
type ProfileResponse struct {
    UserID   int    `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    Email    string `json:"email"`
}