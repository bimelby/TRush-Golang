package models

import "time"

type Alumni struct {
    ID          int       `json:"id"`
    NIM         string    `json:"nim"`
    Nama        string    `json:"nama"`
    Jurusan     string    `json:"jurusan"`
    Angkatan    int       `json:"angkatan"`
    TahunLulus  int       `json:"tahun_lulus"`
    Email       string    `json:"email"`
    NoTelepon   *string   `json:"no_telepon"`
    Alamat      *string   `json:"alamat"`
    UserID      *int      `json:"user_id"`
    IsDeleted   bool      `json:"is_deleted"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
      
}

type CreateAlumniRequest struct {
    NIM        string  `json:"nim" validate:"required"`
    Nama       string  `json:"nama" validate:"required"`
    Jurusan    string  `json:"jurusan" validate:"required"`
    Angkatan   int     `json:"angkatan" validate:"required"`
    TahunLulus int     `json:"tahun_lulus" validate:"required"`
    Email      string  `json:"email" validate:"required,email"`
    NoTelepon  *string `json:"no_telepon"`
    Alamat     *string `json:"alamat"`
}

type UpdateAlumniRequest struct {
    Nama       string  `json:"nama" validate:"required"`
    Jurusan    string  `json:"jurusan" validate:"required"`
    Angkatan   int     `json:"angkatan" validate:"required"`
    TahunLulus int     `json:"tahun_lulus" validate:"required"`
    Email      string  `json:"email" validate:"required,email"`
    NoTelepon  *string `json:"no_telepon"`
    Alamat     *string `json:"alamat"`
}