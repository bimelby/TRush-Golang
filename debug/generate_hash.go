// Buat file debug/generate_hash.go untuk testing
package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    password := "123456"
    
    // Generate hash
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        fmt.Printf("Error generating hash: %v\n", err)
        return
    }
    
    fmt.Printf("Password: %s\n", password)
    fmt.Printf("Hash: %s\n", string(hash))
    
    // Test verify
    err = bcrypt.CompareHashAndPassword(hash, []byte(password))
    if err != nil {
        fmt.Printf("Verification failed: %v\n", err)
    } else {
        fmt.Printf("Verification successful!\n")
    }
}