package repositories

import (
    "alumni-management-system/config"
    "alumni-management-system/models"
    "database/sql"
    "time"
)

type UserRepository interface {
    GetByUsername(username string) (*models.User, string, error) // returns user, password_hash, error
    GetByEmail(email string) (*models.User, string, error)
    GetByID(id int) (*models.User, error)
    Create(user *models.RegisterRequest, passwordHash string) (*models.User, error)
    UpdateLastLogin(userID int) error
}

type userRepository struct {
    db *sql.DB
}

func NewUserRepository() UserRepository {
    return &userRepository{
        db: config.DB,
    }
}

// GetByUsername - ambil user berdasarkan username
func (r *userRepository) GetByUsername(username string) (*models.User, string, error) {
    query := `
        SELECT id, username, email, password_hash, role, created_at, updated_at 
        FROM users 
        WHERE username = $1
    `
    
    var user models.User
    var passwordHash string
    
    row := r.db.QueryRow(query, username)
    err := row.Scan(
        &user.ID, &user.Username, &user.Email, &passwordHash,
        &user.Role, &user.CreatedAt, &user.UpdatedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, "", nil // User not found
        }
        return nil, "", err
    }

    return &user, passwordHash, nil
}

// GetByEmail - ambil user berdasarkan email
func (r *userRepository) GetByEmail(email string) (*models.User, string, error) {
    query := `
        SELECT id, username, email, password_hash, role, created_at, updated_at 
        FROM users 
        WHERE email = $1
    `
    
    var user models.User
    var passwordHash string
    
    row := r.db.QueryRow(query, email)
    err := row.Scan(
        &user.ID, &user.Username, &user.Email, &passwordHash,
        &user.Role, &user.CreatedAt, &user.UpdatedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, "", nil // User not found
        }
        return nil, "", err
    }

    return &user, passwordHash, nil
}

// GetByID - ambil user berdasarkan ID
func (r *userRepository) GetByID(id int) (*models.User, error) {
    query := `
        SELECT id, username, email, role, created_at, updated_at 
        FROM users 
        WHERE id = $1
    `
    
    var user models.User
    row := r.db.QueryRow(query, id)
    
    err := row.Scan(
        &user.ID, &user.Username, &user.Email,
        &user.Role, &user.CreatedAt, &user.UpdatedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }

    return &user, nil
}

// Create - buat user baru (untuk registrasi)
func (r *userRepository) Create(req *models.RegisterRequest, passwordHash string) (*models.User, error) {
    query := `
        INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at, updated_at
    `
    
    now := time.Now()
    var user models.User
    
    err := r.db.QueryRow(
        query, req.Username, req.Email, passwordHash, req.Role, now, now,
    ).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
    
    if err != nil {
        return nil, err
    }

    // Set data dari request
    user.Username = req.Username
    user.Email = req.Email
    user.Role = req.Role

    return &user, nil
}

// UpdateLastLogin - update waktu login terakhir (optional)
func (r *userRepository) UpdateLastLogin(userID int) error {
    query := `UPDATE users SET updated_at = $1 WHERE id = $2`
    _, err := r.db.Exec(query, time.Now(), userID)
    return err
}