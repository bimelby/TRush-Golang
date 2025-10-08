package config

import (
    "database/sql"
    "fmt"
    "log"
    "os"

    _ "github.com/lib/pq"
    "github.com/joho/godotenv"
)

var DB *sql.DB

func ConnectDB() {
    // Load environment variables
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    // Get database configuration from environment
    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    dbname := os.Getenv("DB_NAME")
    sslmode := os.Getenv("DB_SSLMODE")

    // Connection string
    dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        host, port, user, password, dbname, sslmode)

    DB, err = sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal("Gagal koneksi ke database:", err)
    }

    // Test koneksi
    if err = DB.Ping(); err != nil {
        log.Fatal("Gagal ping database:", err)
    }

    fmt.Println("Berhasil terhubung ke database PostgreSQL")
}

func CloseDB() {
    if DB != nil {
        DB.Close()
    }
}