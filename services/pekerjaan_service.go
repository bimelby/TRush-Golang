package services

import (
	"alumni-management-system/models"
	"alumni-management-system/repositories"
	
	"strconv"
	"strings"
    "database/sql"
	

	"github.com/gofiber/fiber/v2"
)

type PekerjaanService interface {
	GetAllPekerjaan(c *fiber.Ctx) error // Signature: func(*fiber.Ctx) error
	GetPekerjaanByID(c *fiber.Ctx) error // Signature: func(*fiber.Ctx) error

	GetPekerjaanByAlumniID(c *fiber.Ctx) error // Signature: func(*fiber.Ctx) error
	CreatePekerjaan(c *fiber.Ctx) error // Signature: func(*fiber.Ctx) error
	UpdatePekerjaan(c *fiber.Ctx) error // Signature: func(*fiber.Ctx) error
	DeletePekerjaan(c *fiber.Ctx) error 
    SoftDeletePekerjaan(c *fiber.Ctx) error// Signature: func(*fiber.Ctx) error
	GetAllNonPekerjaan(c *fiber.Ctx) error
    GetTrashedPekerjaan(c *fiber.Ctx) error
    HardDeleteTrashedPekerjaan(c *fiber.Ctx) error 
    RestoreTrashedPekerjaan(c *fiber.Ctx) error  
      // Signature: func(*fiber.Ctx) error
}

type pekerjaanService struct {
	pekerjaanRepo repositories.PekerjaanRepository
	alumniRepo    repositories.AlumniRepository
}

func NewPekerjaanService(pekerjaanRepo repositories.PekerjaanRepository, alumniRepo repositories.AlumniRepository) PekerjaanService {
	return &pekerjaanService{
		pekerjaanRepo: pekerjaanRepo,
		alumniRepo:    alumniRepo,
	}
}




// SoftDeletePekerjaan - menandai pekerjaan sebagai terhapus secara logis dengan otorisasi
func (s *pekerjaanService) SoftDeletePekerjaan(c *fiber.Ctx) error {
    pekerjaanID, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID pekerjaan tidak valid", "error": err.Error()})
    }

    requesterRole := c.Locals("role").(string)
    requesterUserID := c.Locals("user_id").(int)

    // 1. Dapatkan detail pekerjaan yang akan dihapus
    pekerjaan, err := s.pekerjaanRepo.GetByID(pekerjaanID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal mengambil detail pekerjaan", "error": err.Error()})
    }
    if pekerjaan == nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Pekerjaan tidak ditemukan atau sudah dihapus."})
    }

    // 2. Dapatkan detail alumni yang terkait dengan pekerjaan ini
    alumniPekerjaan, err := s.alumniRepo.GetByID(pekerjaan.AlumniID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal mengambil detail alumni terkait pekerjaan", "error": err.Error()})
    }
    if alumniPekerjaan == nil {
        // Ini seharusnya tidak terjadi jika data konsisten, tapi baik untuk penanganan error
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Alumni terkait pekerjaan tidak ditemukan."})
    }

    // Logika Otorisasi
    if requesterRole == "admin" {
        // Admin boleh menghapus semua riwayat pekerjaan
        err = s.pekerjaanRepo.SoftDelete(pekerjaanID)
        if err != nil {
            if err == sql.ErrNoRows {
                return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Pekerjaan tidak ditemukan atau sudah dihapus."})
            }
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal melakukan soft delete pekerjaan", "error": err.Error()})
        }
        return c.JSON(fiber.Map{"success": true, "message": "Pekerjaan berhasil di-soft delete oleh admin."})
    } else if requesterRole == "user" {
        
        if alumniPekerjaan.UserID == nil || *alumniPekerjaan.UserID != requesterUserID {
            return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Akses ditolak. Anda hanya dapat menghapus pekerjaan Anda sendiri."})
        }

        err = s.pekerjaanRepo.SoftDelete(pekerjaanID)
        if err != nil {
            if err == sql.ErrNoRows {
                return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Pekerjaan tidak ditemukan atau sudah dihapus."})
            }
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal melakukan soft delete pekerjaan", "error": err.Error()})
        }
        return c.JSON(fiber.Map{"success": true, "message": "Pekerjaan Anda berhasil di-soft delete."})
    } else {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Akses ditolak. Role tidak valid."})
    }
}

// GetAllPekerjaan - handle GET /pekerjaan (dengan pagination, search, sorting)
func (s *pekerjaanService) GetAllPekerjaan(c *fiber.Ctx) error {
    page, _ := strconv.Atoi(c.Query("page", "1"))
    limit, _ := strconv.Atoi(c.Query("limit", "10"))
    sortBy := c.Query("sortBy", "id")
    order := c.Query("order", "asc")
    search := c.Query("search", "")

    offset := (page - 1) * limit

    // Validasi input sortBy
    sortByWhitelist := map[string]bool{"id": true, "alumni_id": true, "nama_perusahaan": true, "posisi_jabatan": true, "bidang_industri": true, "lokasi_kerja": true, "tanggal_mulai_kerja": true, "status_pekerjaan": true, "created_at": true}
    if !sortByWhitelist[sortBy] {
        sortBy = "id"
    }
    if strings.ToLower(order) != "desc" {
        order = "asc"
    }

    // Ambil data dari repository
    pekerjaanList, err := s.pekerjaanRepo.GetAllPaginated(search, sortBy, order, limit, offset)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to fetch pekerjaan alumni", "error": err.Error()})
    }

    total, err := s.pekerjaanRepo.CountPekerjaan(search)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to count pekerjaan alumni", "error": err.Error()})
    }

    // Buat response pakai model
    response := models.PekerjaanResponse{
        Data: pekerjaanList,
        Meta: models.MetaInfo{
            Page:   page,
            Limit:  limit,
            Total:  total,
            Pages:  (total + limit - 1) / limit,
            SortBy: sortBy,
            Order:  order,
            Search: search,
        },
    }
    return c.JSON(response)
}

// GetPekerjaanByID - handle GET /pekerjaan/:id
func (s *pekerjaanService) GetPekerjaanByID(c *fiber.Ctx) error {
    id, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid", "error": err.Error()})
    }

    pekerjaan, err := s.pekerjaanRepo.GetByID(id)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to fetch pekerjaan", "error": err.Error()})
    }
    if pekerjaan == nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Pekerjaan tidak ditemukan"})
    }

    return c.JSON(fiber.Map{"success": true, "message": "Data pekerjaan berhasil diambil", "data": pekerjaan})
}

// GetPekerjaanByAlumniID - handle GET /pekerjaan/alumni/:alumni_id
func (s *pekerjaanService) GetPekerjaanByAlumniID(c *fiber.Ctx) error {
    alumniID, err := strconv.Atoi(c.Params("alumni_id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Alumni ID tidak valid", "error": err.Error()})
    }

    // Cek apakah alumni exists
    alumni, err := s.alumniRepo.GetByID(alumniID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to check alumni", "error": err.Error()})
    }

    if alumni == nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Alumni tidak ditemukan"})
    }

    pekerjaanList, err := s.pekerjaanRepo.GetByAlumniID(alumniID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to fetch pekerjaan alumni", "error": err.Error()})
    }

    return c.JSON(fiber.Map{"success": true, "message": "Data pekerjaan alumni berhasil diambil", "data": pekerjaanList})
}

// CreatePekerjaan - handle POST /pekerjaan
func (s *pekerjaanService) CreatePekerjaan(c *fiber.Ctx) error {
    var req models.CreatePekerjaanRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Request body tidak valid", "error": err.Error()})
    }

    // Validasi input
    if req.AlumniID <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Alumni ID harus valid"})
    }

    if req.NamaPerusahaan == "" || req.PosisiJabatan == "" || req.BidangIndustri == "" || req.LokasiKerja == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Nama perusahaan, posisi jabatan, bidang industri, dan lokasi kerja harus diisi"})
    }

    if req.StatusPekerjaan == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Status pekerjaan harus diisi"})
    }

    if req.StatusPekerjaan != "aktif" && req.StatusPekerjaan != "selesai" && req.StatusPekerjaan != "resigned" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Status pekerjaan harus salah satu dari: aktif, selesai, resigned"})
    }

    // Validasi tanggal
    if req.TanggalMulaiKerja.IsZero() {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Tanggal mulai kerja harus diisi"})
    }

    if req.TanggalSelesaiKerja != nil && req.TanggalSelesaiKerja.Before(req.TanggalMulaiKerja) {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Tanggal selesai kerja tidak boleh lebih awal dari tanggal mulai kerja"})
    }

    // Cek apakah alumni exists
    alumni, err := s.alumniRepo.GetByID(req.AlumniID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to check alumni", "error": err.Error()})
    }

    if alumni == nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Alumni tidak ditemukan"})
    }

    pekerjaan, err := s.pekerjaanRepo.Create(&req)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to create pekerjaan", "error": err.Error()})
    }

    return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "message": "Pekerjaan berhasil ditambahkan", "data": pekerjaan})
}

// UpdatePekerjaan - handle PUT /pekerjaan/:id
func (s *pekerjaanService) UpdatePekerjaan(c *fiber.Ctx) error {
    id, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid", "error": err.Error()})
    }

    var req models.UpdatePekerjaanRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Request body tidak valid", "error": err.Error()})
    }

    // Validasi input
    if req.NamaPerusahaan == "" || req.PosisiJabatan == "" || req.BidangIndustri == "" || req.LokasiKerja == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Nama perusahaan, posisi jabatan, bidang industri, dan lokasi kerja harus diisi"})
    }

    if req.StatusPekerjaan == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Status pekerjaan harus diisi"})
    }

    if req.StatusPekerjaan != "aktif" && req.StatusPekerjaan != "selesai" && req.StatusPekerjaan != "resigned" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Status pekerjaan harus salah satu dari: aktif, selesai, resigned"})
    }

    // Validasi tanggal
    if req.TanggalMulaiKerja.IsZero() {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Tanggal mulai kerja harus diisi"})
    }

    if req.TanggalSelesaiKerja != nil && req.TanggalSelesaiKerja.Before(req.TanggalMulaiKerja) {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Tanggal selesai kerja tidak boleh lebih awal dari tanggal mulai kerja"})
    }

    // Cek apakah pekerjaan exists
    existingPekerjaan, err := s.pekerjaanRepo.GetByID(id)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to check pekerjaan", "error": err.Error()})
    }

    if existingPekerjaan == nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Pekerjaan tidak ditemukan"})
    }

    pekerjaan, err := s.pekerjaanRepo.Update(id, &req)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to update pekerjaan", "error": err.Error()})
    }

    return c.JSON(fiber.Map{"success": true, "message": "Pekerjaan berhasil diupdate", "data": pekerjaan})
}

// DeletePekerjaan - handle DELETE /pekerjaan/:id
func (s *pekerjaanService) DeletePekerjaan(c *fiber.Ctx) error {
    id, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid", "error": err.Error()})
    }

    // Cek apakah pekerjaan exists
    existingPekerjaan, err := s.pekerjaanRepo.GetByID(id)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to check pekerjaan", "error": err.Error()})
    }

    if existingPekerjaan == nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Pekerjaan tidak ditemukan"})
    }

    err = s.pekerjaanRepo.Delete(id)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to delete pekerjaan", "error": err.Error()})
    }

    return c.JSON(fiber.Map{"success": true, "message": "Pekerjaan berhasil dihapus"})
}

// GetAllNonPekerjaan - handle GET /pekerjaan/non (jika diperlukan, tanpa pagination)
func (s *pekerjaanService) GetAllNonPekerjaan(c *fiber.Ctx) error {
    pekerjaanList, err := s.pekerjaanRepo.GetAll()
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to fetch pekerjaan alumni", "error": err.Error()})
    }

    return c.JSON(fiber.Map{"success": true, "message": "Data pekerjaan berhasil diambil", "data": pekerjaanList})
}


// GetTrashedPekerjaan - handle GET /pekerjaan/trash (data soft-deleted dengan pagination, search, sorting)
func (s *pekerjaanService) GetTrashedPekerjaan(c *fiber.Ctx) error {
    page, _ := strconv.Atoi(c.Query("page", "1"))
    limit, _ := strconv.Atoi(c.Query("limit", "10"))
    sortBy := c.Query("sortBy", "id")
    order := c.Query("order", "asc")
    search := c.Query("search", "")

    offset := (page - 1) * limit

   
    sortByWhitelist := map[string]bool{"id": true, "alumni_id": true, "nama_perusahaan": true, "posisi_jabatan": true, "bidang_industri": true, "lokasi_kerja": true, "tanggal_mulai_kerja": true, "status_pekerjaan": true, "created_at": true, "updated_at": true}
    if !sortByWhitelist[sortBy] {
        sortBy = "id"
    }
    if strings.ToLower(order) != "desc" {
        order = "asc"
    }

    
    pekerjaanList, err := s.pekerjaanRepo.GetTrashedPaginated(search, sortBy, order, limit, offset)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to fetch trashed pekerjaan", "error": err.Error()})
    }

    total, err := s.pekerjaanRepo.CountTrashed(search)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to count trashed pekerjaan", "error": err.Error()})
    }

    if len(pekerjaanList) == 0 {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Tidak ada data pekerjaan di trash"})
    }

    
    response := models.PekerjaanResponse{
        Data: pekerjaanList,
        Meta: models.MetaInfo{
            Page:   page,
            Limit:  limit,
            Total:  total,
            Pages:  (total + limit - 1) / limit,
            SortBy: sortBy,
            Order:  order,
            Search: search,
        },
    }
    return c.JSON(response)
}



// HardDeleteTrashedPekerjaan - handle DELETE /pekerjaan/trash/:id (hard delete dari trash, hanya admin)
func (s *pekerjaanService) HardDeleteTrashedPekerjaan(c *fiber.Ctx) error {
    id, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "ID tidak valid",
            "error":   err.Error(),
        })
    }

   
    err = s.pekerjaanRepo.HardDeleteTrashed(id)
    if err != nil {
        if err == sql.ErrNoRows {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "success": false,
                "message": "Data pekerjaan tidak ditemukan di trash",
            })
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": "Gagal melakukan hard delete dari trash",
            "error":   err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "success": true,
        "message": "Data pekerjaan berhasil di-hard delete dari trash",
    })
}


// RestoreTrashedPekerjaan - handle PUT /pekerjaan/trash/restore/:id (restore dari trash, hanya admin)
func (s *pekerjaanService) RestoreTrashedPekerjaan(c *fiber.Ctx) error {
    id, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "ID tidak valid",
            "error":   err.Error(),
        })
    }

   
    restored, err := s.pekerjaanRepo.RestoreTrashed(id)
    if err != nil {
        if err == sql.ErrNoRows {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "success": false,
                "message": "Data pekerjaan tidak ditemukan di trash",
            })
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": "Gagal melakukan restore dari trash",
            "error":   err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "success": true,
        "message": "Data pekerjaan berhasil direstore dari trash",
        "data":    restored, 
    })
}
