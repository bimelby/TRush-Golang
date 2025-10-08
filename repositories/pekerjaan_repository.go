package repositories

import (
	"alumni-management-system/config"
	"alumni-management-system/models"
	"database/sql"
	"fmt"
	"time"
)

type PekerjaanRepository interface {
    GetAll() ([]models.PekerjaanAlumni, error)
    GetAllPaginated(search, sortBy, order string, limit, offset int) ([]models.PekerjaanAlumni, error) // New
    CountPekerjaan(search string) (int, error) // New
    GetByID(id int) (*models.PekerjaanAlumni, error)
    GetByAlumniID(alumniID int) ([]models.PekerjaanAlumni, error)
    Create(pekerjaan *models.CreatePekerjaanRequest) (*models.PekerjaanAlumni, error)
    Update(id int, pekerjaan *models.UpdatePekerjaanRequest) (*models.PekerjaanAlumni, error)
    Delete(id int) error
    SoftDelete(id int) error
    GetAllNon() ([]models.PekerjaanAlumni, error)
    GetTrashedPaginated(search, sortBy, order string, limit, offset int) ([]models.PekerjaanAlumni, error) 
    CountTrashed(search string) (int, error)
    HardDeleteTrashed(id int) error
    RestoreTrashed(id int) (*models.PekerjaanAlumni, error) 
}

    
    


type pekerjaanRepository struct {
    db *sql.DB
}

func NewPekerjaanRepository() PekerjaanRepository {
    return &pekerjaanRepository{db: config.DB}
}


func (r *pekerjaanRepository) GetAll() ([]models.PekerjaanAlumni, error) {
    query := `
        SELECT p.id, p.alumni_id, p.nama_perusahaan, p.posisi_jabatan, 
               p.bidang_industri, p.lokasi_kerja, p.gaji_range, 
               p.tanggal_mulai_kerja, p.tanggal_selesai_kerja, 
               p.status_pekerjaan, p.deskripsi_pekerjaan, 
               p.is_deleted, p.created_at, p.updated_at,
               a.nim, a.nama, a.jurusan, a.angkatan, a.tahun_lulus, a.email,a.user_id
        FROM pekerjaan_alumni p
        JOIN alumni a ON p.alumni_id = a.id
        WHERE p.is_deleted = FALSE AND a.is_deleted = FALSE
        ORDER BY p.created_at DESC
    `
    
    rows, err := r.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var pekerjaanList []models.PekerjaanAlumni
    for rows.Next() {
        var pekerjaan models.PekerjaanAlumni
        var alumni models.Alumni
        
        err := rows.Scan(
             &pekerjaan.ID, &pekerjaan.AlumniID, &pekerjaan.NamaPerusahaan,
             &pekerjaan.PosisiJabatan, &pekerjaan.BidangIndustri, &pekerjaan.LokasiKerja,
             &pekerjaan.GajiRange, &pekerjaan.TanggalMulaiKerja, &pekerjaan.TanggalSelesaiKerja,
             &pekerjaan.StatusPekerjaan, &pekerjaan.DeskripsiPekerjaan, &pekerjaan.IsDeleted,
             &pekerjaan.CreatedAt, &pekerjaan.UpdatedAt,
             &alumni.NIM, &alumni.Nama, &alumni.Jurusan, &alumni.Angkatan,
             &alumni.TahunLulus, &alumni.Email, &alumni.UserID,
        )
        if err != nil {
            return nil, err
        }
        
        alumni.ID = pekerjaan.AlumniID
        pekerjaan.Alumni = &alumni
        pekerjaanList = append(pekerjaanList, pekerjaan)
    }

    return pekerjaanList, nil
}
func (r *pekerjaanRepository) SoftDelete(id int) error {
    query := `UPDATE pekerjaan_alumni SET is_deleted = TRUE, updated_at = $1 WHERE id = $2 AND is_deleted = FALSE`
    result, err := r.db.Exec(query, time.Now(), id)
    if err != nil {
        return err
    }
    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return sql.ErrNoRows
    }
    return nil
}

// GetAllPaginated - ambil data pekerjaan dengan pagination, search, dan sorting
func (r *pekerjaanRepository) GetAllPaginated(search, sortBy, order string, limit, offset int) ([]models.PekerjaanAlumni, error) {
    query := fmt.Sprintf(`
        SELECT p.id, p.alumni_id, p.nama_perusahaan, p.posisi_jabatan, 
               p.bidang_industri, p.lokasi_kerja, p.gaji_range, 
               p.tanggal_mulai_kerja, p.tanggal_selesai_kerja, 
               p.status_pekerjaan, p.deskripsi_pekerjaan, 
               p.created_at, p.updated_at,
               a.nim, a.nama, a.jurusan, a.angkatan, a.tahun_lulus, a.email
        FROM pekerjaan_alumni p
        JOIN alumni a ON p.alumni_id = a.id
        WHERE p.nama_perusahaan ILIKE $1 OR p.posisi_jabatan ILIKE $1 OR a.nama ILIKE $1
        ORDER BY %s %s
        LIMIT $2 OFFSET $3
    `, sortBy, order)
    
    rows, err := r.db.Query(query, "%"+search+"%", limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var pekerjaanList []models.PekerjaanAlumni
    for rows.Next() {
        var pekerjaan models.PekerjaanAlumni
        var alumni models.Alumni
        
        err := rows.Scan(
            &pekerjaan.ID, &pekerjaan.AlumniID, &pekerjaan.NamaPerusahaan,
            &pekerjaan.PosisiJabatan, &pekerjaan.BidangIndustri, &pekerjaan.LokasiKerja,
            &pekerjaan.GajiRange, &pekerjaan.TanggalMulaiKerja, &pekerjaan.TanggalSelesaiKerja,
            &pekerjaan.StatusPekerjaan, &pekerjaan.DeskripsiPekerjaan,
            &pekerjaan.CreatedAt, &pekerjaan.UpdatedAt,
            &alumni.NIM, &alumni.Nama, &alumni.Jurusan, &alumni.Angkatan,
            &alumni.TahunLulus, &alumni.Email,
        )
        if err != nil {
            return nil, err
        }
        
        alumni.ID = pekerjaan.AlumniID
        pekerjaan.Alumni = &alumni
        pekerjaanList = append(pekerjaanList, pekerjaan)
    }

    return pekerjaanList, nil
}

// CountPekerjaan - hitung total data pekerjaan untuk pagination
func (r *pekerjaanRepository) CountPekerjaan(search string) (int, error) {
    var total int
    countQuery := `
        SELECT COUNT(*) FROM pekerjaan_alumni p
        JOIN alumni a ON p.alumni_id = a.id
        WHERE p.nama_perusahaan ILIKE $1 OR p.posisi_jabatan ILIKE $1 OR a.nama ILIKE $1
    `
    err := r.db.QueryRow(countQuery, "%"+search+"%").Scan(&total)
    if err != nil && err != sql.ErrNoRows {
        return 0, err
    }
    return total, nil
}


func (r *pekerjaanRepository) GetAllNon() ([]models.PekerjaanAlumni, error) {
    query := `
        SELECT p.id, p.alumni_id, p.nama_perusahaan, p.posisi_jabatan, 
               p.bidang_industri, p.lokasi_kerja, p.gaji_range, 
               p.tanggal_mulai_kerja, p.tanggal_selesai_kerja, 
               p.status_pekerjaan, p.deskripsi_pekerjaan, 
               p.created_at, p.updated_at,
               a.nim, a.nama, a.jurusan, a.angkatan, a.tahun_lulus, a.email
        FROM pekerjaan_alumni p
        JOIN alumni a ON p.alumni_id = a.id
        ORDER BY p.created_at DESC
    `
    
    rows, err := r.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var pekerjaanList []models.PekerjaanAlumni
    for rows.Next() {
        var pekerjaan models.PekerjaanAlumni
        var alumni models.Alumni
        
        err := rows.Scan(
            &pekerjaan.ID, &pekerjaan.AlumniID, &pekerjaan.NamaPerusahaan,
            &pekerjaan.PosisiJabatan, &pekerjaan.BidangIndustri, &pekerjaan.LokasiKerja,
            &pekerjaan.GajiRange, &pekerjaan.TanggalMulaiKerja, &pekerjaan.TanggalSelesaiKerja,
            &pekerjaan.StatusPekerjaan, &pekerjaan.DeskripsiPekerjaan,
            &pekerjaan.CreatedAt, &pekerjaan.UpdatedAt,
            &alumni.NIM, &alumni.Nama, &alumni.Jurusan, &alumni.Angkatan,
            &alumni.TahunLulus, &alumni.Email,
        )
        if err != nil {
            return nil, nil
        }
        
        alumni.ID = pekerjaan.AlumniID
        pekerjaan.Alumni = &alumni
        pekerjaanList = append(pekerjaanList, pekerjaan)
    }

    return pekerjaanList, nil
}


func (r *pekerjaanRepository) GetByID(id int) (*models.PekerjaanAlumni, error) {
    query := `
        SELECT p.id, p.alumni_id, p.nama_perusahaan, p.posisi_jabatan, 
               p.bidang_industri, p.lokasi_kerja, p.gaji_range, 
               p.tanggal_mulai_kerja, p.tanggal_selesai_kerja, 
               p.status_pekerjaan, p.deskripsi_pekerjaan, 
               p.is_deleted, p.created_at, p.updated_at,
               a.nim, a.nama, a.jurusan, a.angkatan, a.tahun_lulus, a.email,a.user_id
        FROM pekerjaan_alumni p
        JOIN alumni a ON p.alumni_id = a.id
        WHERE p.id = $1 AND p.is_deleted = FALSE AND a.is_deleted = FALSE
    `
    
    var pekerjaan models.PekerjaanAlumni
    var alumni models.Alumni
    row := r.db.QueryRow(query, id)
    
    err := row.Scan(
        &pekerjaan.ID, &pekerjaan.AlumniID, &pekerjaan.NamaPerusahaan,
        &pekerjaan.PosisiJabatan, &pekerjaan.BidangIndustri, &pekerjaan.LokasiKerja,
        &pekerjaan.GajiRange, &pekerjaan.TanggalMulaiKerja, &pekerjaan.TanggalSelesaiKerja,
        &pekerjaan.StatusPekerjaan, &pekerjaan.DeskripsiPekerjaan, &pekerjaan.IsDeleted,
        &pekerjaan.CreatedAt, &pekerjaan.UpdatedAt,
        &alumni.NIM, &alumni.Nama, &alumni.Jurusan, &alumni.Angkatan,
        &alumni.TahunLulus, &alumni.Email, &alumni.UserID,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }

    alumni.ID = pekerjaan.AlumniID
    pekerjaan.Alumni = &alumni
    return &pekerjaan, nil
}

func (r *pekerjaanRepository) GetByAlumniID(alumniID int) ([]models.PekerjaanAlumni, error) {
    query := `
        SELECT id, alumni_id, nama_perusahaan, posisi_jabatan, 
               bidang_industri, lokasi_kerja, gaji_range, 
               tanggal_mulai_kerja, tanggal_selesai_kerja, 
               status_pekerjaan, deskripsi_pekerjaan, 
               created_at, updated_at
        FROM pekerjaan_alumni
        WHERE alumni_id = $1
        ORDER BY tanggal_mulai_kerja DESC
    `
    
    rows, err := r.db.Query(query, alumniID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var pekerjaanList []models.PekerjaanAlumni
    for rows.Next() {
        var pekerjaan models.PekerjaanAlumni
        err := rows.Scan(
            &pekerjaan.ID, &pekerjaan.AlumniID, &pekerjaan.NamaPerusahaan,
            &pekerjaan.PosisiJabatan, &pekerjaan.BidangIndustri, &pekerjaan.LokasiKerja,
            &pekerjaan.GajiRange, &pekerjaan.TanggalMulaiKerja, &pekerjaan.TanggalSelesaiKerja,
            &pekerjaan.StatusPekerjaan, &pekerjaan.DeskripsiPekerjaan,
            &pekerjaan.CreatedAt, &pekerjaan.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        pekerjaanList = append(pekerjaanList, pekerjaan)
    }

    return pekerjaanList, nil
}

func (r *pekerjaanRepository) Create(req *models.CreatePekerjaanRequest) (*models.PekerjaanAlumni, error) {
    query := `
        INSERT INTO pekerjaan_alumni (alumni_id, nama_perusahaan, posisi_jabatan, 
                                    bidang_industri, lokasi_kerja, gaji_range, 
                                    tanggal_mulai_kerja, tanggal_selesai_kerja, 
                                    status_pekerjaan, deskripsi_pekerjaan, 
                                    created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        RETURNING id, created_at, updated_at
    `
    
    now := time.Now()
    var pekerjaan models.PekerjaanAlumni
    
    err := r.db.QueryRow(
        query, req.AlumniID, req.NamaPerusahaan, req.PosisiJabatan,
        req.BidangIndustri, req.LokasiKerja, req.GajiRange,
        req.TanggalMulaiKerja, req.TanggalSelesaiKerja, req.StatusPekerjaan,
        req.DeskripsiPekerjaan, now, now,
    ).Scan(&pekerjaan.ID, &pekerjaan.CreatedAt, &pekerjaan.UpdatedAt)
    
    if err != nil {
        return nil, err
    }

    // Set data dari request
    pekerjaan.AlumniID = req.AlumniID
    pekerjaan.NamaPerusahaan = req.NamaPerusahaan
    pekerjaan.PosisiJabatan = req.PosisiJabatan
    pekerjaan.BidangIndustri = req.BidangIndustri
    pekerjaan.LokasiKerja = req.LokasiKerja
    pekerjaan.GajiRange = req.GajiRange
    pekerjaan.TanggalMulaiKerja = req.TanggalMulaiKerja
    pekerjaan.TanggalSelesaiKerja = req.TanggalSelesaiKerja
    pekerjaan.StatusPekerjaan = req.StatusPekerjaan
    pekerjaan.DeskripsiPekerjaan = req.DeskripsiPekerjaan

    return &pekerjaan, nil
}

func (r *pekerjaanRepository) Update(id int, req *models.UpdatePekerjaanRequest) (*models.PekerjaanAlumni, error) {
    query := `
        UPDATE pekerjaan_alumni 
        SET nama_perusahaan = $1, posisi_jabatan = $2, bidang_industri = $3,
            lokasi_kerja = $4, gaji_range = $5, tanggal_mulai_kerja = $6,
            tanggal_selesai_kerja = $7, status_pekerjaan = $8, 
            deskripsi_pekerjaan = $9, updated_at = $10
        WHERE id = $11
    `
    
    now := time.Now()
    result, err := r.db.Exec(
        query, req.NamaPerusahaan, req.PosisiJabatan, req.BidangIndustri,
        req.LokasiKerja, req.GajiRange, req.TanggalMulaiKerja,
        req.TanggalSelesaiKerja, req.StatusPekerjaan, req.DeskripsiPekerjaan,
        now, id,
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

func (r *pekerjaanRepository) Delete(id int) error {
    query := "DELETE FROM pekerjaan_alumni WHERE id = $1"
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
// GetTrashedPaginated - ambil data pekerjaan yang di-soft delete dengan pagination, search, dan sorting
func (r *pekerjaanRepository) GetTrashedPaginated(search, sortBy, order string, limit, offset int) ([]models.PekerjaanAlumni, error) {
    query := fmt.Sprintf(`
        SELECT p.id, p.alumni_id, p.nama_perusahaan, p.posisi_jabatan, 
               p.bidang_industri, p.lokasi_kerja, p.gaji_range, 
               p.tanggal_mulai_kerja, p.tanggal_selesai_kerja, 
               p.status_pekerjaan, p.deskripsi_pekerjaan, 
               p.created_at, p.updated_at,
               a.nim, a.nama, a.jurusan, a.angkatan, a.tahun_lulus, a.email
        FROM pekerjaan_alumni p
        JOIN alumni a ON p.alumni_id = a.id
        WHERE p.is_deleted = TRUE AND (p.nama_perusahaan ILIKE $1 OR p.posisi_jabatan ILIKE $1 OR a.nama ILIKE $1)
        ORDER BY %s %s
        LIMIT $2 OFFSET $3
    `, sortBy, order)
    
    rows, err := r.db.Query(query, "%"+search+"%", limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var pekerjaanList []models.PekerjaanAlumni
    for rows.Next() {
        var pekerjaan models.PekerjaanAlumni
        var alumni models.Alumni
        
        err := rows.Scan(
            &pekerjaan.ID, &pekerjaan.AlumniID, &pekerjaan.NamaPerusahaan,
            &pekerjaan.PosisiJabatan, &pekerjaan.BidangIndustri, &pekerjaan.LokasiKerja,
            &pekerjaan.GajiRange, &pekerjaan.TanggalMulaiKerja, &pekerjaan.TanggalSelesaiKerja,
            &pekerjaan.StatusPekerjaan, &pekerjaan.DeskripsiPekerjaan,
            &pekerjaan.CreatedAt, &pekerjaan.UpdatedAt,
            &alumni.NIM, &alumni.Nama, &alumni.Jurusan, &alumni.Angkatan,
            &alumni.TahunLulus, &alumni.Email,
        )
        if err != nil {
            return nil, err
        }
        
        alumni.ID = pekerjaan.AlumniID
        pekerjaan.Alumni = &alumni
        pekerjaanList = append(pekerjaanList, pekerjaan)
    }

    return pekerjaanList, nil
}

// CountTrashed - hitung total data pekerjaan yang di-soft delete untuk pagination
func (r *pekerjaanRepository) CountTrashed(search string) (int, error) {
    var total int
    countQuery := `
        SELECT COUNT(*) FROM pekerjaan_alumni p
        JOIN alumni a ON p.alumni_id = a.id
        WHERE p.is_deleted = TRUE AND (p.nama_perusahaan ILIKE $1 OR p.posisi_jabatan ILIKE $1 OR a.nama ILIKE $1)
    `
    err := r.db.QueryRow(countQuery, "%"+search+"%").Scan(&total)
    if err != nil && err != sql.ErrNoRows {
        return 0, err
    }
    return total, nil
}


// HardDeleteTrashed - hapus permanen data pekerjaan dari trash (hanya jika is_deleted = TRUE)
func (r *pekerjaanRepository) HardDeleteTrashed(id int) error {
    query := `DELETE FROM pekerjaan_alumni WHERE id = $1 AND is_deleted = TRUE`
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

// RestoreTrashed - kembalikan data pekerjaan dari trash (set is_deleted = FALSE, return data updated)
func (r *pekerjaanRepository) RestoreTrashed(id int) (*models.PekerjaanAlumni, error) {
   
    updateQuery := `UPDATE pekerjaan_alumni SET is_deleted = FALSE, updated_at = $1 WHERE id = $2 AND is_deleted = TRUE`
    now := time.Now()
    result, err := r.db.Exec(updateQuery, now, id)
    if err != nil {
        return nil, err
    }
    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return nil, sql.ErrNoRows 
    }

    
    query := `
        SELECT p.id, p.alumni_id, p.nama_perusahaan, p.posisi_jabatan, 
               p.bidang_industri, p.lokasi_kerja, p.gaji_range, 
               p.tanggal_mulai_kerja, p.tanggal_selesai_kerja, 
               p.status_pekerjaan, p.deskripsi_pekerjaan, 
               p.is_deleted, p.created_at, p.updated_at,
               a.nim, a.nama, a.jurusan, a.angkatan, a.tahun_lulus, a.email, a.user_id
        FROM pekerjaan_alumni p
        JOIN alumni a ON p.alumni_id = a.id
        WHERE p.id = $1
    `
    
    var pekerjaan models.PekerjaanAlumni
    var alumni models.Alumni
    row := r.db.QueryRow(query, id)
    
    err = row.Scan(
        &pekerjaan.ID, &pekerjaan.AlumniID, &pekerjaan.NamaPerusahaan,
        &pekerjaan.PosisiJabatan, &pekerjaan.BidangIndustri, &pekerjaan.LokasiKerja,
        &pekerjaan.GajiRange, &pekerjaan.TanggalMulaiKerja, &pekerjaan.TanggalSelesaiKerja,
        &pekerjaan.StatusPekerjaan, &pekerjaan.DeskripsiPekerjaan, &pekerjaan.IsDeleted,
        &pekerjaan.CreatedAt, &pekerjaan.UpdatedAt,
        &alumni.NIM, &alumni.Nama, &alumni.Jurusan, &alumni.Angkatan,
        &alumni.TahunLulus, &alumni.Email, &alumni.UserID,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }

    alumni.ID = pekerjaan.AlumniID
    pekerjaan.Alumni = &alumni
    return &pekerjaan, nil
}
