package utils

import (
    "alumni-management-system/models"
    "os"
    "strconv"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/joho/godotenv"
)

var jwtSecret []byte

func init() {
    // Load environment variables
    godotenv.Load()
    
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        secret = "default-secret-key-change-in-production-minimum-32-chars"
    }
    jwtSecret = []byte(secret)
}

// Generate JWT token untuk user
func GenerateToken(user models.User) (string, error) {
    // Get expiration time from env or default to 24 hours
    expireHours := 24
    if envHours := os.Getenv("JWT_EXPIRE_HOURS"); envHours != "" {
        if hours, err := strconv.Atoi(envHours); err == nil {
            expireHours = hours
        }
    }

    // Create claims
    claims := models.JWTClaims{
        UserID:   user.ID,
        Username: user.Username,
        Role:     user.Role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "alumni-management-system",
            Subject:   user.Username,
        },
    }

    // Create token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

// Validate JWT token
func ValidateToken(tokenString string) (*models.JWTClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, 
        func(token *jwt.Token) (interface{}, error) {
            // Validate signing method
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrSignatureInvalid
            }
            return jwtSecret, nil
        })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
        return claims, nil
    }

    return nil, jwt.ErrInvalidKey
}

// Extract token from Authorization header
func ExtractTokenFromHeader(authHeader string) string {
    if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
        return authHeader[7:]
    }
    return ""
}