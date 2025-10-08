package repositories

import (
	"alumni-management-system/config"
	"alumni-management-system/models"
	"database/sql"
	"fmt"
	"time"
)

type AlumniRepository interface {
    GetAll() ([]models.Alumni, error)
    GetAllPaginated(search, sortBy, order string, limit, offset int) ([]models.Alumni, error) 
    CountAlumni(search string) (int, error) 
    GetByID(id int) (*models.Alumni, error)
    Create(alumni *models.CreateAlumniRequest) (*models.Alumni, error)
    Update(id int, alumni *models.UpdateAlumniRequest) (*models.Alumni, error)
    Delete(id int) error
    GetAlumniWithoutPekerjaan() ([]models.Alumni, error) 
    GetAlumniByUserID(userID int) (*models.Alumni, error)
    

}

type alumniRepository struct {
    db *sql.DB
}

func NewAlumniRepository() AlumniRepository {
    return &alumniRepository{
        db: config.DB,
    }
}
func (r *alumniRepository) GetAlumniByUserID(userID int) (*models.Alumni, error) {
    query := `
        SELECT id, nim, nama, jurusan, angkatan, tahun_lulus, email,
               no_telepon, alamat, user_id, is_deleted, created_at, updated_at
        FROM alumni
        WHERE user_id = $1 AND is_deleted = FALSE
    `
    var alumni models.Alumni
    row := r.db.QueryRow(query, userID)
    err := row.Scan(
        &alumni.ID, &alumni.NIM, &alumni.Nama, &alumni.Jurusan,
        &alumni.Angkatan, &alumni.TahunLulus, &alumni.Email,
        &alumni.NoTelepon, &alumni.Alamat, &alumni.UserID, &alumni.IsDeleted, &alumni.CreatedAt, &alumni.UpdatedAt,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil // Alumni tidak ditemukan untuk UserID ini
        }
        return nil, err
    }
    return &alumni, nil
}
func (r *alumniRepository) GetAll() ([]models.Alumni, error) {
    query := `
        SELECT id, nim, nama, jurusan, angkatan, tahun_lulus, email, 
               no_telepon, alamat, created_at, updated_at 
        FROM alumni 
        ORDER BY created_at DESC
    `
    
    rows, err := r.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var alumniList []models.Alumni
    for rows.Next() {
        var alumni models.Alumni
        err := rows.Scan(
            &alumni.ID, &alumni.NIM, &alumni.Nama, &alumni.Jurusan,
            &alumni.Angkatan, &alumni.TahunLulus, &alumni.Email,
            &alumni.NoTelepon, &alumni.Alamat, &alumni.CreatedAt, &alumni.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        alumniList = append(alumniList, alumni)
    }

    return alumniList, nil
}


// GetAllPaginated - ambil data alumni dengan pagination, search, dan sorting
func (r *alumniRepository) GetAllPaginated(search, sortBy, order string, limit, offset int) ([]models.Alumni, error) {
    query := fmt.Sprintf(`
        SELECT id, nim, nama, jurusan, angkatan, tahun_lulus, email, 
               no_telepon, alamat, created_at, updated_at 
        FROM alumni 
        WHERE nama ILIKE $1 OR email ILIKE $1 OR nim ILIKE $1 OR jurusan ILIKE $1
        ORDER BY %s %s
        LIMIT $2 OFFSET $3
    `, sortBy, order)
    
    rows, err := r.db.Query(query, "%"+search+"%", limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var alumniList []models.Alumni
    for rows.Next() {
        var alumni models.Alumni
        err := rows.Scan(
            &alumni.ID, &alumni.NIM, &alumni.Nama, &alumni.Jurusan,
            &alumni.Angkatan, &alumni.TahunLulus, &alumni.Email,
            &alumni.NoTelepon, &alumni.Alamat, &alumni.CreatedAt, &alumni.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        alumniList = append(alumniList, alumni)
    }

    return alumniList, nil
}

// CountAlumni - hitung total data alumni untuk pagination
func (r *alumniRepository) CountAlumni(search string) (int, error) {
    var total int
    countQuery := `
        SELECT COUNT(*) FROM alumni 
        WHERE nama ILIKE $1 OR email ILIKE $1 OR nim ILIKE $1 OR jurusan ILIKE $1
    `
    err := r.db.QueryRow(countQuery, "%"+search+"%").Scan(&total)
    if err != nil && err != sql.ErrNoRows {
        return 0, err
    }
    return total, nil
}

func (r *alumniRepository) GetByID(id int) (*models.Alumni, error) {
    query := `
        SELECT id, nim, nama, jurusan, angkatan, tahun_lulus, email, 
               no_telepon, alamat, created_at, updated_at 
        FROM alumni 
        WHERE id = $1
    `
    
    var alumni models.Alumni
    row := r.db.QueryRow(query, id)
    
    err := row.Scan(
        &alumni.ID, &alumni.NIM, &alumni.Nama, &alumni.Jurusan,
        &alumni.Angkatan, &alumni.TahunLulus, &alumni.Email,
        &alumni.NoTelepon, &alumni.Alamat, &alumni.CreatedAt, &alumni.UpdatedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }

    return &alumni, nil
}

func (r *alumniRepository) Create(req *models.CreateAlumniRequest) (*models.Alumni, error) {
    query := `
        INSERT INTO alumni (nim, nama, jurusan, angkatan, tahun_lulus, email, no_telepon, alamat, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id, created_at, updated_at
    `
    
    now := time.Now()
    var alumni models.Alumni
    
    err := r.db.QueryRow(
        query, req.NIM, req.Nama, req.Jurusan, req.Angkatan, req.TahunLulus,
        req.Email, req.NoTelepon, req.Alamat, now, now,
    ).Scan(&alumni.ID, &alumni.CreatedAt, &alumni.UpdatedAt)
    
    if err != nil {
        return nil, err
    }

    // Set data dari request
    alumni.NIM = req.NIM
    alumni.Nama = req.Nama
    alumni.Jurusan = req.Jurusan
    alumni.Angkatan = req.Angkatan
    alumni.TahunLulus = req.TahunLulus
    alumni.Email = req.Email
    alumni.NoTelepon = req.NoTelepon
    alumni.Alamat = req.Alamat

    return &alumni, nil
}

func (r *alumniRepository) Update(id int, req *models.UpdateAlumniRequest) (*models.Alumni, error) {
    query := `
        UPDATE alumni 
        SET nama = $1, jurusan = $2, angkatan = $3, tahun_lulus = $4, 
            email = $5, no_telepon = $6, alamat = $7, updated_at = $8
        WHERE id = $9
    `
    
    now := time.Now()
    result, err := r.db.Exec(
        query, req.Nama, req.Jurusan, req.Angkatan, req.TahunLulus,
        req.Email, req.NoTelepon, req.Alamat, now, id,
    )
    
    if err != nil {
        return nil, err
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return nil, nil
    }

    // Get updated data
    return r.GetByID(id)
}

func (r *alumniRepository) Delete(id int) error {
    query := "DELETE FROM alumni WHERE id = $1"
    result, err := r.db.Exec(query, id)
    if err != nil {
        return err
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return sql.ErrNoRows
    }

    return nil
}

func (r *alumniRepository) GetAlumniWithoutPekerjaan() ([]models.Alumni, error) {
    query := `
        SELECT a.id, a.nim, a.nama, a.jurusan, a.angkatan, a.tahun_lulus, a.email, 
               a.no_telepon, a.alamat, a.created_at, a.updated_at 
        FROM alumni a
        LEFT JOIN pekerjaan_alumni pa ON a.id = pa.alumni_id
        WHERE pa.alumni_id IS NULL
        ORDER BY a.created_at DESC
    `
    
    rows, err := r.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var alumniList []models.Alumni
    for rows.Next() {
        var alumni models.Alumni
        err := rows.Scan(
            &alumni.ID, &alumni.NIM, &alumni.Nama, &alumni.Jurusan,
            &alumni.Angkatan, &alumni.TahunLulus, &alumni.Email,
            &alumni.NoTelepon, &alumni.Alamat, &alumni.CreatedAt, &alumni.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        alumniList = append(alumniList, alumni)
    }
    return alumniList, nil
}