package exam

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Repository -------------------------------------------------------------

// Repository adalah lapisan akses data paket exam. SEMUA SQL ada di sini
// (dipindah verbatim dari exam.go); service/handler tidak menulis SQL
// inline. Transaksi (tx) dijalankan lewat interface querier agar method
// yang sama bisa dipakai dengan pool maupun tx.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// querier adalah subset pgxpool.Pool / pgx.Tx yang dipakai repository.
type querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

func nilOrStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// --- Struktur DB ------------------------------------------------------------
// (Dipindah dari exam.go tanpa perubahan.)

type paket struct {
	ID             string `json:"id"`
	Kode           string `json:"kode"`
	Nama           string `json:"nama"`
	MapelID        string `json:"mapel_id"`
	MapelNama      string `json:"mapel_nama,omitempty"`
	Jenis          string `json:"jenis"`
	DurasiMenit    int    `json:"durasi_menit"`
	JumlahSoal     int    `json:"jumlah_soal"`
	JumlahTersusun int    `json:"jumlah_soal_tersusun,omitempty"`
	AcakSoal       bool   `json:"acak_soal"`
	AcakJawaban    bool   `json:"acak_jawaban"`
	TampilHasil    bool   `json:"tampil_hasil"`
	CreatedAt      any    `json:"created_at"`
}

type jadwal struct {
	ID             string  `json:"id"`
	PaketID        string  `json:"paket_id"`
	PaketKode      string  `json:"paket_kode,omitempty"`
	PaketNama      string  `json:"paket_nama,omitempty"`
	PaketJenis     string  `json:"paket_jenis,omitempty"`
	KelasID        string  `json:"kelas_id"`
	KelasNama      string  `json:"kelas_nama,omitempty"`
	Tanggal        string  `json:"tanggal"`
	JamMulai       string  `json:"jam_mulai"`
	JamSelesai     string  `json:"jam_selesai"`
	Status         string  `json:"status"`
	TokenHash      *string `json:"-"`
	TokenExpiresAt *any    `json:"-"`
	DurasiMenit    int     `json:"durasi_menit,omitempty"`
	AcakSoal       bool    `json:"-"`
	AcakJawaban    bool    `json:"-"`
	TampilHasil    bool    `json:"tampil_hasil,omitempty"`
	CreatedAt      any     `json:"created_at"`
}

type attempt struct {
	ID            string                     `json:"id"`
	JadwalID      string                     `json:"jadwal_id"`
	SiswaID       string                     `json:"siswa_id"`
	Status        string                     `json:"status"`
	MulaiAt       any                        `json:"mulai_at"`
	Deadline      any                        `json:"deadline"`
	SisaWaktuBeku *int                       `json:"sisa_waktu_beku"`
	UrutanSoal    []string                   `json:"-"`
	UrutanOpsi    map[string]json.RawMessage `json:"-"`
	SubmitAt      any                        `json:"submit_at"`
	NilaiAkhir    *float64                   `json:"nilai_akhir"`
	CreatedAt     any                        `json:"created_at"`
}

func scanAttempt(rows pgx.Row) (attempt, error) {
	var a attempt
	var urutanSoal, urutanOpsi []byte
	err := rows.Scan(&a.ID, &a.JadwalID, &a.SiswaID, &a.Status, &a.MulaiAt,
		&a.Deadline, &a.SisaWaktuBeku, &urutanSoal, &urutanOpsi,
		&a.SubmitAt, &a.NilaiAkhir, &a.CreatedAt)
	if err != nil {
		return a, err
	}
	_ = json.Unmarshal(urutanSoal, &a.UrutanSoal)
	_ = json.Unmarshal(urutanOpsi, &a.UrutanOpsi)
	if a.UrutanOpsi == nil {
		a.UrutanOpsi = map[string]json.RawMessage{}
	}
	return a, nil
}

// GetJadwal memuat satu jadwal + info paket/kelas. (Dulu: getJadwal.)
func (r *Repository) GetJadwal(ctx context.Context, id string) (jadwal, error) {
	var j jadwal
	var tokenExpAny any
	err := r.pool.QueryRow(ctx,
		`SELECT j.id, j.paket_id, j.kelas_id, j.tanggal::text, j.jam_mulai::text,
		        j.jam_selesai::text, j.status, j.token_hash, j.token_expires_at, j.created_at,
		        p.kode, p.nama, p.jenis, p.durasi_menit, p.acak_soal, p.acak_jawaban, p.tampil_hasil,
		        COALESCE(k.nama, '')
		 FROM jadwal_ujian j
		 JOIN paket_ujian p ON p.id = j.paket_id
		 LEFT JOIN kelas k ON k.id = j.kelas_id
		 WHERE j.id = $1`, id).
		Scan(&j.ID, &j.PaketID, &j.KelasID, &j.Tanggal, &j.JamMulai, &j.JamSelesai,
			&j.Status, &j.TokenHash, &tokenExpAny, &j.CreatedAt,
			&j.PaketKode, &j.PaketNama, &j.PaketJenis, &j.DurasiMenit,
			&j.AcakSoal, &j.AcakJawaban, &j.TampilHasil, &j.KelasNama)
	if err != nil {
		return j, err
	}
	if tokenExpAny != nil {
		j.TokenExpiresAt = &tokenExpAny
	}
	return j, nil
}

// GetAttempt memuat satu attempt. (Dulu: getAttempt.)
func (r *Repository) GetAttempt(ctx context.Context, id string) (attempt, error) {
	return scanAttempt(r.pool.QueryRow(ctx,
		`SELECT id, jadwal_id, siswa_id, status, mulai_at, deadline,
		        sisa_waktu_beku, urutan_soal, urutan_opsi, submit_at,
		        nilai_akhir, created_at
		 FROM attempt WHERE id = $1`, id))
}

// JadwalStatus membaca status jadwal sebuah attempt. Bila gagal dibaca,
// kembalikan "pause" (aman: auto-submit dilewati saat ragu).
// (Dulu: jadwalStatusUntuk.)
func (r *Repository) JadwalStatus(ctx context.Context, jadwalID string) string {
	var st string
	if err := r.pool.QueryRow(ctx,
		`SELECT status FROM jadwal_ujian WHERE id = $1`, jadwalID).Scan(&st); err != nil {
		return "pause"
	}
	return st
}

// --- Settings (dibaca toleran: value jsonb maupun text) --------------------
// (Dipindah dari exam.go tanpa perubahan SQL.)

func (r *Repository) settingRaw(ctx context.Context, key string) (string, bool) {
	var raw []byte
	err := r.pool.QueryRow(ctx, `SELECT value FROM settings WHERE key = $1`, key).Scan(&raw)
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(raw)), true
}

func (r *Repository) settingInt(ctx context.Context, key string, def int) int {
	raw, ok := r.settingRaw(ctx, key)
	if !ok {
		return def
	}
	// Bentuk JSON number ("900") atau text ("900") atau '"900"'.
	var f float64
	if err := json.Unmarshal([]byte(raw), &f); err == nil {
		return int(f)
	}
	if n, err := strconv.Atoi(strings.Trim(raw, `"`)); err == nil {
		return n
	}
	return def
}

func (r *Repository) settingBool(ctx context.Context, key string, def bool) bool {
	raw, ok := r.settingRaw(ctx, key)
	if !ok {
		return def
	}
	var b bool
	if err := json.Unmarshal([]byte(raw), &b); err == nil {
		return b
	}
	if parsed, err := strconv.ParseBool(strings.Trim(raw, `"`)); err == nil {
		return parsed
	}
	return def
}

// TokenExpiryDetik membaca settings.token_expiry_detik (default 900).
func (r *Repository) TokenExpiryDetik(ctx context.Context) int {
	return r.settingInt(ctx, "token_expiry_detik", 900)
}

// ToleransiTelatMenit membaca settings.toleransi_telat_menit (default 15).
func (r *Repository) ToleransiTelatMenit(ctx context.Context) int {
	return r.settingInt(ctx, "toleransi_telat_menit", 15)
}

// MaxPelanggaran membaca settings.max_pelanggaran (default 2, min 1).
func (r *Repository) MaxPelanggaran(ctx context.Context) int {
	if v := r.settingInt(ctx, "max_pelanggaran", 2); v >= 1 {
		return v
	}
	return 2
}

// TampilHasilEfektif = paket.tampil_hasil AND settings.tampil_hasil_ke_siswa.
func (r *Repository) TampilHasilEfektif(ctx context.Context, paketTampil bool) bool {
	if !paketTampil {
		return false
	}
	return r.settingBool(ctx, "tampil_hasil_ke_siswa", false)
}

// CatatAudit menulis satu baris audit log. Gagal catat TIDAK menggagalkan
// operasi utama. (Dulu: auditCatat; menulis langsung ke audit_log karena
// paket audit Tim 5 belum compile di Echo v5 — lihat currentUser.)
func (r *Repository) CatatAudit(ctx context.Context, actorID, aksi, entitas, entitasID string, detail any, ip string) {
	detailJSON, err := json.Marshal(detail)
	if err != nil {
		detailJSON = []byte("{}")
	}
	var actor any
	if strings.TrimSpace(actorID) != "" {
		actor = actorID
	}
	_, _ = r.pool.Exec(ctx,
		`INSERT INTO audit_log (actor_id, aksi, entitas, entitas_id, detail, ip)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		actor, aksi, entitas, entitasID, detailJSON, ip)
}
