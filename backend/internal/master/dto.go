package master

import (
	"encoding/json"
	"time"
)

// ---------------------------------------------------------------------
// DTO
// ---------------------------------------------------------------------

type jurusanDTO struct {
	ID        string    `json:"id"`
	Kode      string    `json:"kode"`
	Nama      string    `json:"nama"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type kelasDTO struct {
	ID          string    `json:"id"`
	JurusanID   string    `json:"jurusan_id"`
	Tingkat     int32     `json:"tingkat"`
	Nama        string    `json:"nama"`
	TahunAjaran string    `json:"tahun_ajaran"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type mapelDTO struct {
	ID           string    `json:"id"`
	Kode         string    `json:"kode"`
	Nama         string    `json:"nama"`
	Jenis        string    `json:"jenis"`
	JurusanID    *string   `json:"jurusan_id"`
	KelasTingkat *int32    `json:"kelas_tingkat"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type guruDTO struct {
	ID        string          `json:"id"`
	UserID    *string         `json:"user_id"`
	NIP       *string         `json:"nip"`
	Nama      string          `json:"nama"`
	MapelIDs  json.RawMessage `json:"mapel_ids"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type siswaDTO struct {
	ID         string    `json:"id"`
	UserID     *string   `json:"user_id"`
	NISN       string    `json:"nisn"`
	NIS        *string   `json:"nis"`
	Nama       string    `json:"nama"`
	KelasID    string    `json:"kelas_id"`
	JurusanID  string    `json:"jurusan_id"`
	TahunMasuk *int32    `json:"tahun_masuk"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type kompetensiDTO struct {
	ID        string    `json:"id"`
	MapelID   string    `json:"mapel_id"`
	Kode      string    `json:"kode"`
	Deskripsi string    `json:"deskripsi"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type kisiKisiDTO struct {
	ID           string    `json:"id"`
	MapelID      string    `json:"mapel_id"`
	KelasTingkat int32     `json:"kelas_tingkat"`
	Kelompok     string    `json:"kelompok"`
	KompetensiID *string   `json:"kompetensi_id"`
	JumlahSoal   int32     `json:"jumlah_soal"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ---------------------------------------------------------------------
// Import Siswa dari CSV
// ---------------------------------------------------------------------

type importRowError struct {
	Baris int    `json:"baris"`
	Pesan string `json:"pesan"`
}

// importResult: SudahAda = baris yang NISN-nya sudah terdaftar di DB —
// dihitung TERPISAH dari gagal (Fase 3, BUG-02: import IDEMPOTEN.
// Retry import yang sama tidak lagi melaporkan "gagal" massal;
// import kedua 300 siswa yang sama -> sukses 0, sudah_ada 300, gagal 0).
type importResult struct {
	Sukses   int              `json:"sukses"`
	Gagal    int              `json:"gagal"`
	SudahAda int              `json:"sudah_ada"`
	Errors   []importRowError `json:"errors"`
	Catatan  string           `json:"catatan,omitempty"`
}

const (
	maxImportRows   = 2000
	maxImportFileMB = 5
	kelasLookupByID   = `SELECT id, jurusan_id FROM kelas WHERE id = $1`
	kelasLookupByNama = `SELECT id, jurusan_id FROM kelas WHERE LOWER(nama) = LOWER($1) ORDER BY created_at DESC LIMIT 1`
)
