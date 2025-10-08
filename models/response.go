package models

// MetaInfo - informasi pagination & filter
type MetaInfo struct {
    Page   int    `json:"page"`
    Limit  int    `json:"limit"`
    Total  int    `json:"total"`
    Pages  int    `json:"pages"`
    SortBy string `json:"sortBy"`
    Order  string `json:"order"`
    Search string `json:"search"`
}

// AlumniResponse - hasil akhir untuk endpoint /alumni
type AlumniResponse struct {
    Data []Alumni `json:"data"`
    Meta MetaInfo `json:"meta"`
}

// PekerjaanResponse - hasil akhir untuk endpoint /pekerjaan
type PekerjaanResponse struct {
    Data []PekerjaanAlumni `json:"data"`
    Meta MetaInfo          `json:"meta"`
}
