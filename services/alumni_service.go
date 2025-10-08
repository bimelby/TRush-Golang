package services

import (
	"alumni-management-system/models"
	"alumni-management-system/repositories"

	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AlumniService interface {
	GetAllAlumni(c *fiber.Ctx) error              // Signature: func(*fiber.Ctx) error
	GetAlumniByID(c *fiber.Ctx) error             // Signature: func(*fiber.Ctx) error
	CreateAlumni(c *fiber.Ctx) error              // Signature: func(*fiber.Ctx) error
	UpdateAlumni(c *fiber.Ctx) error              // Signature: func(*fiber.Ctx) error
	DeleteAlumni(c *fiber.Ctx) error              // Signature: func(*fiber.Ctx) error
	GetAlumniWithoutPekerjaan(c *fiber.Ctx) error // New: Get alumni without jobs
}

type alumniService struct {
	alumniRepo repositories.AlumniRepository
}

func NewAlumniService(alumniRepo repositories.AlumniRepository) AlumniService {
	return &alumniService{
		alumniRepo: alumniRepo,
	}
}

// GetAllAlumni - handle GET /alumni (dengan pagination, search, sorting)
func (s *alumniService) GetAllAlumni(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	sortBy := c.Query("sortBy", "id")
	order := c.Query("order", "asc")
	search := c.Query("search", "")

	offset := (page - 1) * limit

	// Validasi input sortBy
	sortByWhitelist := map[string]bool{"id": true, "nim": true, "nama": true, "jurusan": true, "angkatan": true, "tahun_lulus": true, "email": true, "created_at": true}
	if !sortByWhitelist[sortBy] {
		sortBy = "id"
	}
	if strings.ToLower(order) != "desc" {
		order = "asc"
	}

	// Ambil data dari repository
	alumniList, err := s.alumniRepo.GetAllPaginated(search, sortBy, order, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to fetch alumni", "error": err.Error()})
	}

	total, err := s.alumniRepo.CountAlumni(search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to count alumni", "error": err.Error()})
	}

	// Buat response pakai model
	response := models.AlumniResponse{
		Data: alumniList,
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

// GetAlumniByID - handle GET /alumni/:id
func (s *alumniService) GetAlumniByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid", "error": err.Error()})
	}

	alumni, err := s.alumniRepo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to fetch alumni", "error": err.Error()})
	}
	if alumni == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Alumni tidak ditemukan"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Data alumni berhasil diambil", "data": alumni})
}

// CreateAlumni - handle POST /alumni
func (s *alumniService) CreateAlumni(c *fiber.Ctx) error {
	var req models.CreateAlumniRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Request body tidak valid", "error": err.Error()})
	}

	// Validasi input
	if req.NIM == "" || req.Nama == "" || req.Jurusan == "" || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "NIM, nama, jurusan, dan email harus diisi"})
	}

	if req.Angkatan <= 0 || req.TahunLulus <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Angkatan dan tahun lulus harus valid"})
	}

	if req.TahunLulus < req.Angkatan {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Tahun lulus tidak boleh lebih kecil dari angkatan"})
	}

	alumni, err := s.alumniRepo.Create(&req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to create alumni", "error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "message": "Alumni berhasil ditambahkan", "data": alumni})
}

// UpdateAlumni - handle PUT /alumni/:id
func (s *alumniService) UpdateAlumni(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid", "error": err.Error()})
	}

	var req models.UpdateAlumniRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Request body tidak valid", "error": err.Error()})
	}

	// Validasi input
	if req.Nama == "" || req.Jurusan == "" || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Nama, jurusan, dan email harus diisi"})
	}

	if req.Angkatan <= 0 || req.TahunLulus <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Angkatan dan tahun lulus harus valid"})
	}

	if req.TahunLulus < req.Angkatan {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Tahun lulus tidak boleh lebih kecil dari angkatan"})
	}

	// Cek apakah alumni exists
	existingAlumni, err := s.alumniRepo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to check alumni", "error": err.Error()})
	}

	if existingAlumni == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Alumni tidak ditemukan"})
	}

	alumni, err := s.alumniRepo.Update(id, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to update alumni", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Alumni berhasil diupdate", "data": alumni})
}

// DeleteAlumni - handle DELETE /alumni/:id
func (s *alumniService) DeleteAlumni(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid", "error": err.Error()})
	}

	// Cek apakah alumni exists
	existingAlumni, err := s.alumniRepo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to check alumni", "error": err.Error()})
	}

	if existingAlumni == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Alumni tidak ditemukan"})
	}

	err = s.alumniRepo.Delete(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to delete alumni", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Alumni berhasil dihapus"})
}

func (s *alumniService) GetAlumniWithoutPekerjaan(c *fiber.Ctx) error {
	alumniList, err := s.alumniRepo.GetAlumniWithoutPekerjaan()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to fetch alumni without jobs", "error": err.Error()})
	}
	if len(alumniList) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Tidak ada alumni yang belum memiliki pekerjaan"})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Data alumni tanpa pekerjaan berhasil diambil", "data": alumniList})
}
