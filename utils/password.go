package utils

import (
    "golang.org/x/crypto/bcrypt"
)

// Hash password menggunakan bcrypt
func HashPassword(password string) (string, error) {
    // Generate hash dengan default cost (10)
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// Verify password dengan hash
func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// Generate hash untuk password baru (utility function)
func GeneratePasswordHash(password string) (string, error) {
    return HashPassword(password)
}

// Validate password strength (optional)
func ValidatePasswordStrength(password string) bool {
    // Minimum 6 characters
    if len(password) < 6 {
        return false
    }
    
    return true
}