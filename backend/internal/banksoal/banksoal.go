package banksoal

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

func writeError(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

func writeOK(c *echo.Context, data any) error {
	return c.JSON(http.StatusOK, map[string]any{"data": data})
}

func writeCreated(c *echo.Context, data any) error {
	return c.JSON(http.StatusCreated, map[string]any{"data": data})
}

// requirePerm: 401 bila belum login (context auth Tim 1 kosong),
// 403 bila tidak punya permission. Role admin lolos semua permission
// (lihat audit.HasPermission).
func requirePerm(c *echo.Context, perm string) (userID, role string, ok bool) {
	uid, r, _, _, ok := audit.CurrentUser(c)
	if !ok {
		_ = writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
		return "", "", false
	}
	if !audit.HasPermission(c, perm) {
		_ = writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin: "+perm+".")
		return "", "", false
	}
	return uid, r, true
}

// forbidSoalUntukProktor: proktor & pengawas TIDAK boleh melihat isi
// soal (README + seed: proktor hanya punya banksoal.read untuk metadata
// bank; pengawas tidak punya permission banksoal sama sekali).
// Dipanggil di SEMUA endpoint soal (list/get/create/update/delete/import).
func forbidSoalUntukProktor(c *echo.Context, role string) bool {
	if role == "proktor" {
		_ = writeError(c, http.StatusForbidden, "forbidden", "Proktor tidak boleh melihat isi soal.")
		return false
	}
	return true
}

func isValidUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, ch := range s {
		switch i {
		case 8, 13, 18, 23:
			if ch != '-' {
				return false
			}
		default:
			if !(ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F') {
				return false
			}
		}
	}
	return true
}

func parsePagination(c *echo.Context) (limit, offset int) {
	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset, err = strconv.Atoi(c.QueryParam("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}
	return limit, offset
}

func strPtr(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// --- Model JSON ---

type bankRow struct {
	ID           string  `json:"id"`
	Kode         string  `json:"kode"`
	Nama         string  `json:"nama"`
	MapelID      string  `json:"mapel_id"`
	KelasTingkat int     `json:"kelas_tingkat"`
	GuruID       *string `json:"guru_id"`
	IsLocked     bool    `json:"is_locked"`
	JumlahSoal   int     `json:"jumlah_soal"`
	CreatedAt    any     `json:"created_at"`
	UpdatedAt    any     `json:"updated_at"`
}

type soalRow struct {
	ID           string          `json:"id"`
	BankID       string          `json:"bank_id"`
	Tipe         string          `json:"tipe"`
	KisiKisiID   *string         `json:"kisi_kisi_id"`
	KompetensiID *string         `json:"kompetensi_id"`
	Stimulus     string          `json:"stimulus"`
	Pertanyaan   string          `json:"pertanyaan"`
	MediaURL     string          `json:"media_url"`
	Jawaban      json.RawMessage `json:"jawaban"`
	Kunci        json.RawMessage `json:"kunci,omitempty"`
	// Pembahasan (Fase 3, kolom soal.pembahasan migrasi 0011):
	// penjelasan guru; HANYA untuk guru/admin via endpoint bank soal.
	// TIDAK PERNAH dikirim ke siswa sebelum submit (endpoint attempt
	// tidak memilih kolom ini; hasil hanya setelah submit + tampil_hasil).
	Pembahasan   string          `json:"pembahasan"`
	Bobot        int             `json:"bobot"`
	Level        string          `json:"level"`
	Status       string          `json:"status"`
	CreatedAt    any             `json:"created_at"`
	UpdatedAt    any             `json:"updated_at"`
}

// loadBank mengambil bank; mengembalikan (bank, ditemukan, err).
func loadBank(c *echo.Context, pool *pgxpool.Pool, id string) (*bankRow, bool, error) {
	if !isValidUUID(id) {
		return nil, false, nil
	}
	var b bankRow
	err := pool.QueryRow(c.Request().Context(),
		`SELECT b.id, b.kode, b.nama, b.mapel_id, b.kelas_tingkat, b.guru_id,
		        b.is_locked, b.created_at, b.updated_at,
		        (SELECT COUNT(*) FROM soal s WHERE s.bank_id = b.id)
		 FROM bank_soal b WHERE b.id = $1`, id).
		Scan(&b.ID, &b.Kode, &b.Nama, &b.MapelID, &b.KelasTingkat, &b.GuruID,
			&b.IsLocked, &b.CreatedAt, &b.UpdatedAt, &b.JumlahSoal)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &b, true, nil
}

// checkGuruScope: guru hanya boleh memutasi bank miliknya (guru_id =
// dirinya) atau bank tanpa pemilik. Admin bebas.
func checkGuruScope(c *echo.Context, role, userID string, b *bankRow) bool {
	if role == "admin" {
		return true
	}
	if role == "guru" {
		if b.GuruID == nil || *b.GuruID == "" || *b.GuruID == userID {
			return true
		}
		_ = writeError(c, http.StatusForbidden, "forbidden", "Bank soal ini milik guru lain.")
		return false
	}
	return true
}

// bankDipakaiPaket: true bila ada soal dari bank ini yang tersusun di
// paket manapun (via paket_soal JOIN soal). Kueri langsung ke tabel
// kanonis 0004 — tanpa workaround information_schema.
func bankDipakaiPaket(c *echo.Context, pool *pgxpool.Pool, bankID string) (bool, error) {
	ctx := c.Request().Context()
	var n int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM paket_soal ps
		 JOIN soal s ON s.id = ps.soal_id
		 WHERE s.bank_id = $1`, bankID).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

// --- Input ---

type bankInput struct {
	Kode         string  `json:"kode"`
	Nama         string  `json:"nama"`
	MapelID      string  `json:"mapel_id"`
	KelasTingkat int     `json:"kelas_tingkat"`
	GuruID       *string `json:"guru_id"`
}

type soalInput struct {
	Tipe         string          `json:"tipe"`
	KisiKisiID   *string         `json:"kisi_kisi_id"`
	KompetensiID *string         `json:"kompetensi_id"`
	Stimulus     string          `json:"stimulus"`
	Pertanyaan   string          `json:"pertanyaan"`
	MediaURL     string          `json:"media_url"`
	Jawaban      json.RawMessage `json:"jawaban"`
	Kunci        json.RawMessage `json:"kunci"`
	Pembahasan   string          `json:"pembahasan"`
	Bobot        int             `json:"bobot"`
	Level        string          `json:"level"`
	Status       string          `json:"status"`
}

func (in *soalInput) validate() error {
	if !ValidTipe(in.Tipe) {
		return fmt.Errorf("tipe %q tidak dikenal", in.Tipe)
	}
	if strings.TrimSpace(in.Pertanyaan) == "" {
		return fmt.Errorf("pertanyaan wajib diisi")
	}
	if in.Tipe == "studi_kasus" && strings.TrimSpace(in.Stimulus) == "" {
		return fmt.Errorf("stimulus wajib diisi untuk tipe studi_kasus")
	}
	if in.Bobot <= 0 {
		return fmt.Errorf("bobot harus > 0")
	}
	if in.Level != "" && !ValidLevel(in.Level) {
		return fmt.Errorf("level harus mudah / sedang / sulit")
	}
	if in.Status != "" {
		switch in.Status {
		case "draft", "aktif", "nonaktif", "perlu_revisi", "diarsipkan":
		default:
			return fmt.Errorf("status %q tidak dikenal", in.Status)
		}
	}
	if in.Tipe == "teks_rumpang" {
		// Jumlah rumpang (___) harus sama dengan jumlah jawaban kunci.
		var kunci map[string]any
		if err := json.Unmarshal(in.Kunci, &kunci); err == nil {
			if ks, ok := asStringSlice(kunci["jawaban"]); ok {
				if strings.Count(in.Pertanyaan, "___") != len(ks) {
					return fmt.Errorf("jumlah rumpang ___ (%d) harus sama dengan jumlah jawaban kunci (%d)",
						strings.Count(in.Pertanyaan, "___"), len(ks))
				}
			}
		}
	}
	return ValidateSoal(in.Tipe, in.Jawaban, in.Kunci)
}

// Register mendaftarkan endpoint bank soal (Tim 3).
//
//	GET    /api/v1/bank-soal                        (banksoal.read)
//	POST   /api/v1/bank-soal                        (banksoal.create)
//	GET    /api/v1/bank-soal/:id                    (banksoal.read)
//	PUT    /api/v1/bank-soal/:id                    (banksoal.update)
//	DELETE /api/v1/bank-soal/:id                    (banksoal.delete)
//	POST   /api/v1/bank-soal/:id/lock               (banksoal.lock)
//	POST   /api/v1/bank-soal/:id/unlock             (banksoal.lock)
//	POST   /api/v1/bank-soal/import                 (banksoal.import)
//	GET    /api/v1/bank-soal/:bankId/soal           (banksoal.read, guru/admin saja)
//	POST   /api/v1/bank-soal/:bankId/soal           (banksoal.create)
//	GET    /api/v1/bank-soal/:bankId/soal/:soalId   (banksoal.read, guru/admin saja)
//	PUT    /api/v1/bank-soal/:bankId/soal/:soalId   (banksoal.update)
//	DELETE /api/v1/bank-soal/:bankId/soal/:soalId   (banksoal.delete)
//
// Aturan kunci:
//   - Bank terkunci -> mutasi soal (tambah/ubah/hapus/import) = 409.
//   - Hapus bank -> 409 bila terkunci atau dipakai paket.
//   - Kunci jawaban hanya untuk guru/admin; proktor tidak melihat isi
//     soal; pengawas/siswa/penguji UKK tidak punya akses (403).
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	// ---- Bank soal ----

	e.GET("/api/v1/bank-soal", func(c *echo.Context) error {
		if _, _, ok := requirePerm(c, "banksoal.read"); !ok {
			return nil
		}
		limit, offset := parsePagination(c)
		search := strPtr(c.QueryParam("q"))
		mapelID := strPtr(c.QueryParam("mapel_id"))
		var kelas *int
		if v := c.QueryParam("kelas_tingkat"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				kelas = &n
			}
		}
		var locked *bool
		if v := c.QueryParam("is_locked"); v != "" {
			if b, err := strconv.ParseBool(v); err == nil {
				locked = &b
			}
		}
		var kelasArg, lockedArg any
		if kelas != nil {
			kelasArg = *kelas
		}
		if locked != nil {
			lockedArg = *locked
		}
		rows, err := pool.Query(c.Request().Context(),
			`SELECT b.id, b.kode, b.nama, b.mapel_id, b.kelas_tingkat, b.guru_id,
			        b.is_locked, b.created_at, b.updated_at,
			        (SELECT COUNT(*) FROM soal s WHERE s.bank_id = b.id)
			 FROM bank_soal b
			 WHERE ($1::text IS NULL OR b.kode ILIKE '%' || $1 || '%' OR b.nama ILIKE '%' || $1 || '%')
			   AND ($2::uuid IS NULL OR b.mapel_id = $2)
			   AND ($3::int IS NULL OR b.kelas_tingkat = $3)
			   AND ($4::boolean IS NULL OR b.is_locked = $4)
			 ORDER BY b.created_at DESC LIMIT $5 OFFSET $6`,
			search, mapelID, kelasArg, lockedArg, limit, offset)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca bank soal.")
		}
		defer rows.Close()
		items := []bankRow{}
		for rows.Next() {
			var b bankRow
			if err := rows.Scan(&b.ID, &b.Kode, &b.Nama, &b.MapelID, &b.KelasTingkat,
				&b.GuruID, &b.IsLocked, &b.CreatedAt, &b.UpdatedAt, &b.JumlahSoal); err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca bank soal.")
			}
			items = append(items, b)
		}
		var total int
		_ = pool.QueryRow(c.Request().Context(),
			`SELECT COUNT(*) FROM bank_soal b
			 WHERE ($1::text IS NULL OR b.kode ILIKE '%' || $1 || '%' OR b.nama ILIKE '%' || $1 || '%')
			   AND ($2::uuid IS NULL OR b.mapel_id = $2)
			   AND ($3::int IS NULL OR b.kelas_tingkat = $3)
			   AND ($4::boolean IS NULL OR b.is_locked = $4)`,
			search, mapelID, kelasArg, lockedArg).Scan(&total)
		return c.JSON(http.StatusOK, map[string]any{
			"data": map[string]any{"items": items, "total": total, "limit": limit, "offset": offset},
		})
	})

	e.POST("/api/v1/bank-soal", func(c *echo.Context) error {
		userID, role, ok := requirePerm(c, "banksoal.create")
		if !ok {
			return nil
		}
		var in bankInput
		if err := c.Bind(&in); err != nil {
			return writeError(c, http.StatusBadRequest, "invalid_body", "Body JSON tidak valid.")
		}
		if strings.TrimSpace(in.Kode) == "" || strings.TrimSpace(in.Nama) == "" {
			return writeError(c, http.StatusBadRequest, "validation_error", "kode dan nama bank wajib diisi.")
		}
		if !isValidUUID(in.MapelID) {
			return writeError(c, http.StatusBadRequest, "validation_error", "mapel_id wajib UUID mapel yang valid.")
		}
		if in.KelasTingkat != 10 && in.KelasTingkat != 11 && in.KelasTingkat != 12 {
			return writeError(c, http.StatusBadRequest, "validation_error", "kelas_tingkat harus 10, 11, atau 12.")
		}
		// Guru yang membuat bank menjadi pemiliknya; admin boleh menetapkan guru_id.
		var guruID any
		if role == "guru" {
			guruID = userID
		} else if in.GuruID != nil && *in.GuruID != "" {
			if !isValidUUID(*in.GuruID) {
				return writeError(c, http.StatusBadRequest, "validation_error", "guru_id harus UUID yang valid.")
			}
			guruID = *in.GuruID
		}
		var b bankRow
		err := pool.QueryRow(c.Request().Context(),
			`INSERT INTO bank_soal (kode, nama, mapel_id, kelas_tingkat, guru_id)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id, kode, nama, mapel_id, kelas_tingkat, guru_id, is_locked, created_at, updated_at, 0`,
			in.Kode, in.Nama, in.MapelID, in.KelasTingkat, guruID).
			Scan(&b.ID, &b.Kode, &b.Nama, &b.MapelID, &b.KelasTingkat, &b.GuruID,
				&b.IsLocked, &b.CreatedAt, &b.UpdatedAt, &b.JumlahSoal)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
				return writeError(c, http.StatusConflict, "duplicate_kode", "kode bank soal sudah dipakai.")
			}
			if strings.Contains(err.Error(), "foreign key") || strings.Contains(err.Error(), "violates") {
				return writeError(c, http.StatusBadRequest, "validation_error", "mapel_id / guru_id tidak ditemukan.")
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membuat bank soal.")
		}
		_ = audit.Catat(c.Request().Context(), pool, userID, "banksoal.create", "bank_soal", b.ID,
			map[string]any{"kode": b.Kode, "nama": b.Nama}, c.RealIP())
		return writeCreated(c, b)
	})

	e.GET("/api/v1/bank-soal/:id", func(c *echo.Context) error {
		if _, _, ok := requirePerm(c, "banksoal.read"); !ok {
			return nil
		}
		b, found, err := loadBank(c, pool, c.Param("id"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca bank soal.")
		}
		if !found {
			return writeError(c, http.StatusNotFound, "not_found", "Bank soal tidak ditemukan.")
		}
		return writeOK(c, b)
	})

	e.PUT("/api/v1/bank-soal/:id", func(c *echo.Context) error {
		userID, role, ok := requirePerm(c, "banksoal.update")
		if !ok {
			return nil
		}
		b, found, err := loadBank(c, pool, c.Param("id"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca bank soal.")
		}
		if !found {
			return writeError(c, http.StatusNotFound, "not_found", "Bank soal tidak ditemukan.")
		}
		if !checkGuruScope(c, role, userID, b) {
			return nil
		}
		var in bankInput
		if err := c.Bind(&in); err != nil {
			return writeError(c, http.StatusBadRequest, "invalid_body", "Body JSON tidak valid.")
		}
		if strings.TrimSpace(in.Kode) == "" || strings.TrimSpace(in.Nama) == "" {
			return writeError(c, http.StatusBadRequest, "validation_error", "kode dan nama bank wajib diisi.")
		}
		if !isValidUUID(in.MapelID) {
			return writeError(c, http.StatusBadRequest, "validation_error", "mapel_id wajib UUID mapel yang valid.")
		}
		if in.KelasTingkat != 10 && in.KelasTingkat != 11 && in.KelasTingkat != 12 {
			return writeError(c, http.StatusBadRequest, "validation_error", "kelas_tingkat harus 10, 11, atau 12.")
		}
		// guru_id hanya boleh diubah admin; guru tidak bisa memindahkan kepemilikan.
		guruID := b.GuruID
		if role == "admin" && in.GuruID != nil {
			if *in.GuruID == "" {
				guruID = nil
			} else if isValidUUID(*in.GuruID) {
				guruID = in.GuruID
			} else {
				return writeError(c, http.StatusBadRequest, "validation_error", "guru_id harus UUID yang valid.")
			}
		}
		var guruArg any
		if guruID != nil {
			guruArg = *guruID
		}
		_, err = pool.Exec(c.Request().Context(),
			`UPDATE bank_soal SET kode=$2, nama=$3, mapel_id=$4, kelas_tingkat=$5,
			 guru_id=$6, updated_at=now() WHERE id=$1`,
			b.ID, in.Kode, in.Nama, in.MapelID, in.KelasTingkat, guruArg)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
				return writeError(c, http.StatusConflict, "duplicate_kode", "kode bank soal sudah dipakai.")
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengubah bank soal.")
		}
		_ = audit.Catat(c.Request().Context(), pool, userID, "banksoal.update", "bank_soal", b.ID,
			map[string]any{"kode": in.Kode, "nama": in.Nama}, c.RealIP())
		updated, _, _ := loadBank(c, pool, b.ID)
		return writeOK(c, updated)
	})

	e.DELETE("/api/v1/bank-soal/:id", func(c *echo.Context) error {
		userID, role, ok := requirePerm(c, "banksoal.delete")
		if !ok {
			return nil
		}
		b, found, err := loadBank(c, pool, c.Param("id"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca bank soal.")
		}
		if !found {
			return writeError(c, http.StatusNotFound, "not_found", "Bank soal tidak ditemukan.")
		}
		if !checkGuruScope(c, role, userID, b) {
			return nil
		}
		if b.IsLocked {
			return writeError(c, http.StatusConflict, "bank_locked", "Bank soal terkunci — tidak bisa dihapus. Buka kunci dulu (oleh admin / bukan pembuat).")
		}
		inUse, err := bankDipakaiPaket(c, pool, b.ID)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa pemakaian bank.")
		}
		if inUse {
			return writeError(c, http.StatusConflict, "bank_in_use", "Bank soal dipakai paket ujian — tidak bisa dihapus.")
		}
		if _, err := pool.Exec(c.Request().Context(), `DELETE FROM bank_soal WHERE id=$1`, b.ID); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghapus bank soal.")
		}
		_ = audit.Catat(c.Request().Context(), pool, userID, "banksoal.delete", "bank_soal", b.ID,
			map[string]any{"kode": b.Kode, "nama": b.Nama}, c.RealIP())
		return writeOK(c, map[string]any{"deleted": true, "id": b.ID})
	})

	// Lock / unlock. SoD-1 (ANALISIS Bab 11.4): pembuat bank TIDAK BOLEH
	// mengunci/membuka bank buatannya sendiri — pengunci harus orang
	// kedua (guru lain) atau admin.
	setLock := func(locked bool) echo.HandlerFunc {
		return func(c *echo.Context) error {
			userID, role, ok := requirePerm(c, "banksoal.lock")
			if !ok {
				return nil
			}
			b, found, err := loadBank(c, pool, c.Param("id"))
			if err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca bank soal.")
			}
			if !found {
				return writeError(c, http.StatusNotFound, "not_found", "Bank soal tidak ditemukan.")
			}
			if role == "guru" && b.GuruID != nil && *b.GuruID == userID {
				return writeError(c, http.StatusForbidden, "sod_violation",
					"Pemisahan tugas (SoD-1): pembuat bank tidak boleh mengunci/membuka bank buatannya sendiri. Minta guru lain atau admin.")
			}
			if b.IsLocked == locked {
				return writeOK(c, b)
			}
			if _, err := pool.Exec(c.Request().Context(),
				`UPDATE bank_soal SET is_locked=$2, updated_at=now() WHERE id=$1`, b.ID, locked); err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengubah kunci bank.")
			}
			aksi := "banksoal.lock"
			if !locked {
				aksi = "banksoal.unlock"
			}
			_ = audit.Catat(c.Request().Context(), pool, userID, aksi, "bank_soal", b.ID,
				map[string]any{"kode": b.Kode, "is_locked": locked}, c.RealIP())
			updated, _, _ := loadBank(c, pool, b.ID)
			return writeOK(c, updated)
		}
	}
	e.POST("/api/v1/bank-soal/:id/lock", setLock(true))
	e.POST("/api/v1/bank-soal/:id/unlock", setLock(false))

	// ---- Import CSV ----
	// POST /api/v1/bank-soal/import (didaftarkan sebelum rute :id agar
	// segmen statis "import" tidak tertukar dengan parameter :id).
	e.POST("/api/v1/bank-soal/import", func(c *echo.Context) error {
		userID, role, ok := requirePerm(c, "banksoal.import")
		if !ok {
			return nil
		}
		if !forbidSoalUntukProktor(c, role) {
			return nil
		}
		var body struct {
			BankID  string `json:"bank_id"`
			CSVText string `json:"csv_text"`
			CSV     string `json:"csv"`
			Mode    string `json:"mode"` // all_or_nothing (default) | skip_errors
		}
		if err := c.Bind(&body); err != nil {
			return writeError(c, http.StatusBadRequest, "invalid_body", "Body JSON tidak valid.")
		}
		csvText := body.CSVText
		if csvText == "" {
			csvText = body.CSV
		}
		if !isValidUUID(body.BankID) {
			return writeError(c, http.StatusBadRequest, "validation_error", "bank_id wajib UUID bank yang valid.")
		}
		if strings.TrimSpace(csvText) == "" {
			return writeError(c, http.StatusBadRequest, "validation_error", "csv_text wajib diisi (teks CSV).")
		}
		mode := strings.ToLower(strings.TrimSpace(body.Mode))
		if mode == "" {
			mode = "all_or_nothing"
		}
		if mode != "all_or_nothing" && mode != "skip_errors" {
			return writeError(c, http.StatusBadRequest, "validation_error", "mode harus all_or_nothing atau skip_errors.")
		}
		b, found, err := loadBank(c, pool, body.BankID)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca bank soal.")
		}
		if !found {
			return writeError(c, http.StatusNotFound, "not_found", "Bank soal tidak ditemukan.")
		}
		if !checkGuruScope(c, role, userID, b) {
			return nil
		}
		if b.IsLocked {
			return writeError(c, http.StatusConflict, "bank_locked", "Bank soal terkunci — import ditolak.")
		}
		parsed, importErrs := parseCSVImport(csvText)
		if len(parsed) == 0 && len(importErrs) == 0 {
			return writeError(c, http.StatusBadRequest, "validation_error", "CSV tidak berisi baris soal.")
		}
		if mode == "all_or_nothing" && len(importErrs) > 0 {
			return writeOK(c, map[string]any{
				"sukses": 0, "gagal": len(importErrs), "errors": importErrs,
				"mode": mode, "disimpan": false,
			})
		}
		sukses := 0
		for _, row := range parsed {
			_, err := pool.Exec(c.Request().Context(),
				`INSERT INTO soal (bank_id, tipe, stimulus, pertanyaan, media_url,
				 jawaban, kunci, pembahasan, bobot, level, status)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'draft')`,
				b.ID, row.tipe, row.stimulus, row.pertanyaan, "",
				row.jawaban, row.kunci, row.pembahasan, row.bobot, row.level)
			if err != nil {
				importErrs = append(importErrs, map[string]any{
					"baris": row.baris, "pesan": "Gagal menyimpan ke database.",
				})
				continue
			}
			sukses++
		}
		_ = audit.Catat(c.Request().Context(), pool, userID, "banksoal.import", "bank_soal", b.ID,
			map[string]any{"sukses": sukses, "gagal": len(importErrs), "mode": mode}, c.RealIP())
		return writeOK(c, map[string]any{
			"sukses": sukses, "gagal": len(importErrs), "errors": importErrs,
			"mode": mode, "disimpan": sukses > 0,
		})
	})

	// ---- Soal ----
	loadSoal := func(c *echo.Context, bankID, soalID string) (*soalRow, bool, error) {
		if !isValidUUID(soalID) {
			return nil, false, nil
		}
		var s soalRow
		err := pool.QueryRow(c.Request().Context(),
			`SELECT id, bank_id, tipe, kisi_kisi_id, kompetensi_id, stimulus, pertanyaan,
			        media_url, jawaban, kunci, pembahasan, bobot, level, status, created_at, updated_at
			 FROM soal WHERE id=$1 AND bank_id=$2`, soalID, bankID).
			Scan(&s.ID, &s.BankID, &s.Tipe, &s.KisiKisiID, &s.KompetensiID, &s.Stimulus,
				&s.Pertanyaan, &s.MediaURL, &s.Jawaban, &s.Kunci, &s.Pembahasan, &s.Bobot, &s.Level,
				&s.Status, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			if err.Error() == "no rows in result set" {
				return nil, false, nil
			}
			return nil, false, err
		}
		return &s, true, nil
	}

	e.GET("/api/v1/bank-soal/:bankId/soal", func(c *echo.Context) error {
		_, role, ok := requirePerm(c, "banksoal.read")
		if !ok {
			return nil
		}
		if !forbidSoalUntukProktor(c, role) {
			return nil
		}
		b, found, err := loadBank(c, pool, c.Param("bankId"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca bank soal.")
		}
		if !found {
			return writeError(c, http.StatusNotFound, "not_found", "Bank soal tidak ditemukan.")
		}
		limit, offset := parsePagination(c)
		tipe := strPtr(c.QueryParam("tipe"))
		level := strPtr(c.QueryParam("level"))
		kisi := strPtr(c.QueryParam("kisi_kisi_id"))
		search := strPtr(c.QueryParam("q"))
		rows, err := pool.Query(c.Request().Context(),
			`SELECT id, bank_id, tipe, kisi_kisi_id, kompetensi_id, stimulus, pertanyaan,
			        media_url, jawaban, kunci, pembahasan, bobot, level, status, created_at, updated_at
			 FROM soal WHERE bank_id=$1
			   AND ($2::text IS NULL OR tipe=$2)
			   AND ($3::text IS NULL OR level=$3)
			   AND ($4::uuid IS NULL OR kisi_kisi_id=$4)
			   AND ($5::text IS NULL OR pertanyaan ILIKE '%' || $5 || '%')
			 ORDER BY created_at ASC LIMIT $6 OFFSET $7`,
			b.ID, tipe, level, kisi, search, limit, offset)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca soal.")
		}
		defer rows.Close()
		items := []soalRow{}
		for rows.Next() {
			var s soalRow
			if err := rows.Scan(&s.ID, &s.BankID, &s.Tipe, &s.KisiKisiID, &s.KompetensiID,
				&s.Stimulus, &s.Pertanyaan, &s.MediaURL, &s.Jawaban, &s.Kunci, &s.Pembahasan, &s.Bobot,
				&s.Level, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca soal.")
			}
			items = append(items, s)
		}
		var total int
		_ = pool.QueryRow(c.Request().Context(),
			`SELECT COUNT(*) FROM soal WHERE bank_id=$1
			   AND ($2::text IS NULL OR tipe=$2)
			   AND ($3::text IS NULL OR level=$3)
			   AND ($4::uuid IS NULL OR kisi_kisi_id=$4)
			   AND ($5::text IS NULL OR pertanyaan ILIKE '%' || $5 || '%')`,
			b.ID, tipe, level, kisi, search).Scan(&total)
		return c.JSON(http.StatusOK, map[string]any{
			"data": map[string]any{"items": items, "total": total, "limit": limit, "offset": offset},
		})
	})

	e.POST("/api/v1/bank-soal/:bankId/soal", func(c *echo.Context) error {
		userID, role, ok := requirePerm(c, "banksoal.create")
		if !ok {
			return nil
		}
		if !forbidSoalUntukProktor(c, role) {
			return nil
		}
		b, found, err := loadBank(c, pool, c.Param("bankId"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca bank soal.")
		}
		if !found {
			return writeError(c, http.StatusNotFound, "not_found", "Bank soal tidak ditemukan.")
		}
		if !checkGuruScope(c, role, userID, b) {
			return nil
		}
		if b.IsLocked {
			return writeError(c, http.StatusConflict, "bank_locked", "Bank soal terkunci — soal tidak bisa ditambah.")
		}
		var in soalInput
		if err := c.Bind(&in); err != nil {
			return writeError(c, http.StatusBadRequest, "invalid_body", "Body JSON tidak valid.")
		}
		if err := in.validate(); err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", err.Error())
		}
		level := NormalLevel(in.Level)
		if level == "" {
			level = "sedang"
		}
		status := in.Status
		if status == "" {
			status = "aktif"
		}
		var s soalRow
		err = pool.QueryRow(c.Request().Context(),
			`INSERT INTO soal (bank_id, tipe, kisi_kisi_id, kompetensi_id, stimulus,
			 pertanyaan, media_url, jawaban, kunci, pembahasan, bobot, level, status)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
			 RETURNING id, bank_id, tipe, kisi_kisi_id, kompetensi_id, stimulus,
			 pertanyaan, media_url, jawaban, kunci, pembahasan, bobot, level, status, created_at, updated_at`,
			b.ID, in.Tipe, strPtr(ptrVal(in.KisiKisiID)), strPtr(ptrVal(in.KompetensiID)),
			in.Stimulus, in.Pertanyaan, in.MediaURL, in.Jawaban, in.Kunci, in.Pembahasan, in.Bobot, level, status).
			Scan(&s.ID, &s.BankID, &s.Tipe, &s.KisiKisiID, &s.KompetensiID, &s.Stimulus,
				&s.Pertanyaan, &s.MediaURL, &s.Jawaban, &s.Kunci, &s.Pembahasan, &s.Bobot, &s.Level,
				&s.Status, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membuat soal.")
		}
		_ = audit.Catat(c.Request().Context(), pool, userID, "banksoal.create", "soal", s.ID,
			map[string]any{"bank_id": b.ID, "tipe": s.Tipe}, c.RealIP())
		return writeCreated(c, s)
	})

	e.GET("/api/v1/bank-soal/:bankId/soal/:soalId", func(c *echo.Context) error {
		_, role, ok := requirePerm(c, "banksoal.read")
		if !ok {
			return nil
		}
		if !forbidSoalUntukProktor(c, role) {
			return nil
		}
		s, found, err := loadSoal(c, c.Param("bankId"), c.Param("soalId"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca soal.")
		}
		if !found {
			return writeError(c, http.StatusNotFound, "not_found", "Soal tidak ditemukan.")
		}
		return writeOK(c, s)
	})

	e.PUT("/api/v1/bank-soal/:bankId/soal/:soalId", func(c *echo.Context) error {
		userID, role, ok := requirePerm(c, "banksoal.update")
		if !ok {
			return nil
		}
		if !forbidSoalUntukProktor(c, role) {
			return nil
		}
		b, found, err := loadBank(c, pool, c.Param("bankId"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca bank soal.")
		}
		if !found {
			return writeError(c, http.StatusNotFound, "not_found", "Bank soal tidak ditemukan.")
		}
		if !checkGuruScope(c, role, userID, b) {
			return nil
		}
		if b.IsLocked {
			return writeError(c, http.StatusConflict, "bank_locked", "Bank soal terkunci — soal tidak bisa diubah.")
		}
		s, found, err := loadSoal(c, b.ID, c.Param("soalId"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca soal.")
		}
		if !found {
			return writeError(c, http.StatusNotFound, "not_found", "Soal tidak ditemukan.")
		}
		var in soalInput
		if err := c.Bind(&in); err != nil {
			return writeError(c, http.StatusBadRequest, "invalid_body", "Body JSON tidak valid.")
		}
		if err := in.validate(); err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", err.Error())
		}
		level := NormalLevel(in.Level)
		if level == "" {
			level = "sedang"
		}
		status := in.Status
		if status == "" {
			status = s.Status
		}
		_, err = pool.Exec(c.Request().Context(),
			`UPDATE soal SET tipe=$3, kisi_kisi_id=$4, kompetensi_id=$5, stimulus=$6,
			 pertanyaan=$7, media_url=$8, jawaban=$9, kunci=$10, bobot=$11,
			 level=$12, status=$13, pembahasan=$14, updated_at=now() WHERE id=$1 AND bank_id=$2`,
			s.ID, b.ID, in.Tipe, strPtr(ptrVal(in.KisiKisiID)), strPtr(ptrVal(in.KompetensiID)),
			in.Stimulus, in.Pertanyaan, in.MediaURL, in.Jawaban, in.Kunci, in.Bobot, level, status, in.Pembahasan)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengubah soal.")
		}
		_ = audit.Catat(c.Request().Context(), pool, userID, "banksoal.update", "soal", s.ID,
			map[string]any{"bank_id": b.ID, "tipe": in.Tipe}, c.RealIP())
		updated, _, _ := loadSoal(c, b.ID, s.ID)
		return writeOK(c, updated)
	})

	e.DELETE("/api/v1/bank-soal/:bankId/soal/:soalId", func(c *echo.Context) error {
		userID, role, ok := requirePerm(c, "banksoal.delete")
		if !ok {
			return nil
		}
		if !forbidSoalUntukProktor(c, role) {
			return nil
		}
		b, found, err := loadBank(c, pool, c.Param("bankId"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca bank soal.")
		}
		if !found {
			return writeError(c, http.StatusNotFound, "not_found", "Bank soal tidak ditemukan.")
		}
		if !checkGuruScope(c, role, userID, b) {
			return nil
		}
		if b.IsLocked {
			return writeError(c, http.StatusConflict, "bank_locked", "Bank soal terkunci — soal tidak bisa dihapus.")
		}
		s, found, err := loadSoal(c, b.ID, c.Param("soalId"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca soal.")
		}
		if !found {
			return writeError(c, http.StatusNotFound, "not_found", "Soal tidak ditemukan.")
		}
		if _, err := pool.Exec(c.Request().Context(),
			`DELETE FROM soal WHERE id=$1 AND bank_id=$2`, s.ID, b.ID); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghapus soal.")
		}
		_ = audit.Catat(c.Request().Context(), pool, userID, "banksoal.delete", "soal", s.ID,
			map[string]any{"bank_id": b.ID, "tipe": s.Tipe}, c.RealIP())
		return writeOK(c, map[string]any{"deleted": true, "id": s.ID})
	})
}

func ptrVal(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// --- Import CSV ---

type csvSoal struct {
	baris      int
	tipe       string
	pertanyaan string
	stimulus   string
	jawaban    json.RawMessage
	kunci      json.RawMessage
	pembahasan string
	bobot      int
	level      string
	// Nilai mentah dari parser Word (.docx) — gabungkanDocx memindahkannya
	// ke excelSoal: media_url tersimpan, kompetensi dipetakan ke ID via
	// lookup kode, kisi_kisi ditolak E30 (sama seperti jalur .xlsx).
	// Jalur CSV tidak mengisi keempatnya (CSV tidak punya kolomnya).
	// Catatan: excelSoal juga punya field bernama sama — selektor
	// es.mediaURL dll. menunjuk field excelSoal (depth 0 menaungi
	// field embedded depth 1), jadi tidak ada ambiguitas.
	mediaURL   string
	mediaAlt   string
	kisiKisi   string
	kompetensi string
}

var csvColumns = []string{
	"tipe", "pertanyaan", "opsi_a", "opsi_b", "opsi_c", "opsi_d", "opsi_e",
	"kunci", "bobot", "level", "stimulus",
}

// parseCSVImport mem-parsing teks CSV menjadi soal yang sudah divalidasi.
// Kolom: tipe,pertanyaan,opsi_a,opsi_b,opsi_c,opsi_d,opsi_e,kunci,bobot,level,stimulus
// Maks 500 baris data. Baris yang gagal dikembalikan di errors dan TIDAK
// disimpan (pemanggil memutuskan all_or_nothing vs skip_errors).
func parseCSVImport(csvText string) ([]csvSoal, []map[string]any) {
	r := csv.NewReader(strings.NewReader(csvText))
	r.TrimLeadingSpace = true
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err != nil {
		return nil, []map[string]any{{"baris": 0, "pesan": "CSV tidak bisa dibaca: " + err.Error()}}
	}
	if len(records) < 2 {
		return nil, []map[string]any{{"baris": 0, "pesan": "CSV harus punya header + minimal 1 baris soal."}}
	}
	header := map[string]int{}
	for i, h := range records[0] {
		header[strings.ToLower(strings.TrimSpace(h))] = i
	}
	var errors []map[string]any
	fail := func(baris int, msg string) {
		errors = append(errors, map[string]any{"baris": baris, "pesan": msg})
	}
	for _, col := range []string{"tipe", "pertanyaan", "kunci"} {
		if _, ok := header[col]; !ok {
			return nil, []map[string]any{{"baris": 1, "pesan": "Kolom wajib hilang di header: " + col}}
		}
	}
	dataRows := records[1:]
	if len(dataRows) > 500 {
		return nil, []map[string]any{{"baris": 0, "pesan": "CSV melebihi batas 500 baris soal."}}
	}
	get := func(rec []string, col string) string {
		idx, ok := header[col]
		if !ok || idx >= len(rec) {
			return ""
		}
		return strings.TrimSpace(rec[idx])
	}
	var out []csvSoal
	for i, rec := range dataRows {
		baris := i + 2 // header = baris 1
		// Lewati baris kosong.
		allEmpty := true
		for _, v := range rec {
			if strings.TrimSpace(v) != "" {
				allEmpty = false
			}
		}
		if allEmpty {
			continue
		}
		tipe := strings.ToLower(strings.TrimSpace(get(rec, "tipe")))
		pertanyaan := strings.TrimSpace(get(rec, "pertanyaan"))
		kunciStr := strings.TrimSpace(get(rec, "kunci"))
		stimulus := strings.TrimSpace(get(rec, "stimulus"))
		if !ValidTipe(tipe) {
			fail(baris, fmt.Sprintf("tipe %q tidak dikenal", tipe))
			continue
		}
		if pertanyaan == "" {
			fail(baris, "pertanyaan wajib diisi")
			continue
		}
		if kunciStr == "" {
			fail(baris, "kunci wajib diisi")
			continue
		}
		bobot := 1
		if raw := get(rec, "bobot"); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil || n <= 0 {
				fail(baris, fmt.Sprintf("bobot %q harus bilangan bulat > 0", raw))
				continue
			}
			bobot = n
		}
		level := NormalLevel(get(rec, "level"))
		if level == "" {
			level = "sedang"
		}
		if !ValidLevel(level) {
			fail(baris, fmt.Sprintf("level %q harus mudah / sedang / sulit", level))
			continue
		}
		jawaban, kunci, err := buildCSVSoal(tipe, rec, get, kunciStr, stimulus)
		if err != nil {
			fail(baris, err.Error())
			continue
		}
		if err := ValidateSoal(tipe, jawaban, kunci); err != nil {
			fail(baris, err.Error())
			continue
		}
		if tipe == "teks_rumpang" {
			var k map[string]any
			_ = json.Unmarshal(kunci, &k)
			if ks, ok := asStringSlice(k["jawaban"]); ok && strings.Count(pertanyaan, "___") != len(ks) {
				fail(baris, "jumlah rumpang ___ harus sama dengan jumlah jawaban kunci (pisah dengan |)")
				continue
			}
		}
		if tipe == "studi_kasus" && stimulus == "" {
			fail(baris, "stimulus wajib diisi untuk tipe studi_kasus")
			continue
		}
		out = append(out, csvSoal{
			baris: baris, tipe: tipe, pertanyaan: pertanyaan, stimulus: stimulus,
			jawaban: jawaban, kunci: kunci, pembahasan: strings.TrimSpace(get(rec, "pembahasan")),
			bobot: bobot, level: level,
		})
	}
	return out, errors
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func opsiFromCSV(get func([]string, string) string, rec []string) []map[string]any {
	labels := []string{"A", "B", "C", "D", "E"}
	var out []map[string]any
	for i, lbl := range labels {
		teks := strings.TrimSpace(get(rec, "opsi_"+strings.ToLower(lbl)))
		_ = i
		if teks != "" {
			out = append(out, map[string]any{"label": lbl, "teks": teks})
		}
	}
	return out
}

func splitAndTrim(s, sep string) []string {
	var out []string
	for _, part := range strings.Split(s, sep) {
		if t := strings.TrimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// buildCSVSoal membangun jawaban/kunci JSON dari satu baris CSV.
func buildCSVSoal(tipe string, rec []string, get func([]string, string) string, kunciStr, stimulus string) (json.RawMessage, json.RawMessage, error) {
	switch tipe {
	case "pg", "studi_kasus", "pilihan_gambar":
		opsi := opsiFromCSV(get, rec)
		if len(opsi) < 2 {
			return nil, nil, fmt.Errorf("opsi_a..opsi_e: minimal 2 opsi terisi")
		}
		k := strings.ToUpper(kunciStr)
		found := false
		for _, o := range opsi {
			if o["label"] == k {
				found = true
			}
		}
		if !found {
			return nil, nil, fmt.Errorf("kunci %q tidak cocok dengan label opsi yang terisi", kunciStr)
		}
		return mustJSON(map[string]any{"opsi": opsi}),
			mustJSON(map[string]any{"opsi": k}), nil
	case "pg_kompleks", "pg_multi":
		opsi := opsiFromCSV(get, rec)
		if len(opsi) < 2 {
			return nil, nil, fmt.Errorf("opsi_a..opsi_e: minimal 2 opsi terisi")
		}
		ks := splitAndTrim(strings.ToUpper(kunciStr), ",")
		labels := map[string]bool{}
		for _, o := range opsi {
			labels[o["label"].(string)] = true
		}
		for _, k := range ks {
			if !labels[k] {
				return nil, nil, fmt.Errorf("kunci %q tidak cocok dengan label opsi yang terisi", k)
			}
		}
		if len(ks) == 0 {
			return nil, nil, fmt.Errorf("kunci wajib minimal satu label opsi")
		}
		return mustJSON(map[string]any{"opsi": opsi}),
			mustJSON(map[string]any{"opsi": ks}), nil
	case "benar_salah":
		opsi := opsiFromCSV(get, rec)
		if len(opsi) < 2 {
			return nil, nil, fmt.Errorf("opsi_a..opsi_e: isi pernyataan (minimal 2)")
		}
		pernyataan := make([]string, 0, len(opsi))
		for _, o := range opsi {
			pernyataan = append(pernyataan, o["teks"].(string))
		}
		parts := splitAndTrim(kunciStr, ",")
		if len(parts) != len(pernyataan) {
			return nil, nil, fmt.Errorf("kunci harus %d nilai B/S, dipisah koma", len(pernyataan))
		}
		var nilai []bool
		for _, p := range parts {
			switch strings.ToLower(p) {
			case "b", "benar", "true", "1":
				nilai = append(nilai, true)
			case "s", "salah", "false", "0":
				nilai = append(nilai, false)
			default:
				return nil, nil, fmt.Errorf("kunci %q harus B atau S", p)
			}
		}
		return mustJSON(map[string]any{"pernyataan": pernyataan}),
			mustJSON(map[string]any{"nilai": nilai}), nil
	case "menjodohkan", "peta_konsep":
		opsi := opsiFromCSV(get, rec)
		if len(opsi) < 2 {
			return nil, nil, fmt.Errorf("opsi_a..opsi_e: isi pasangan format kiri::kanan (minimal 2)")
		}
		var pas []map[string]any
		for _, o := range opsi {
			kv := strings.SplitN(o["teks"].(string), "::", 2)
			if len(kv) != 2 || strings.TrimSpace(kv[0]) == "" || strings.TrimSpace(kv[1]) == "" {
				return nil, nil, fmt.Errorf("opsi %s harus format kiri::kanan", o["label"])
			}
			pas = append(pas, map[string]any{"kiri": strings.TrimSpace(kv[0]), "kanan": strings.TrimSpace(kv[1])})
		}
		var kPas []map[string]any
		for _, part := range strings.Split(kunciStr, ";") {
			kv := strings.SplitN(strings.TrimSpace(part), "::", 2)
			if len(kv) != 2 || strings.TrimSpace(kv[0]) == "" || strings.TrimSpace(kv[1]) == "" {
				return nil, nil, fmt.Errorf("kunci harus pasangan format kiri::kanan dipisah ;")
			}
			kPas = append(kPas, map[string]any{"kiri": strings.TrimSpace(kv[0]), "kanan": strings.TrimSpace(kv[1])})
		}
		return mustJSON(map[string]any{"pasangan": pas}),
			mustJSON(map[string]any{"pasangan": kPas}), nil
	case "isian":
		alts := splitAndTrim(kunciStr, "|")
		if len(alts) == 0 {
			return nil, nil, fmt.Errorf("kunci isian wajib diisi (alternatif dipisah |)")
		}
		return mustJSON(map[string]any{}),
			mustJSON(map[string]any{"jawaban": alts}), nil
	case "uraian", "upload_file", "praktik_singkat":
		// Kolom kunci CSV = rubrik.
		return mustJSON(map[string]any{}),
			mustJSON(map[string]any{"rubrik": kunciStr}), nil
	case "urutkan":
		opsi := opsiFromCSV(get, rec)
		if len(opsi) < 2 {
			return nil, nil, fmt.Errorf("opsi_a..opsi_e: isi butir yang diurutkan (minimal 2)")
		}
		items := make([]string, 0, len(opsi))
		labelToIdx := map[string]int{}
		for i, o := range opsi {
			items = append(items, o["teks"].(string))
			labelToIdx[o["label"].(string)] = i
		}
		ks := splitAndTrim(strings.ToUpper(kunciStr), ",")
		if len(ks) != len(items) {
			return nil, nil, fmt.Errorf("kunci harus %d label opsi (urutan benar), dipisah koma", len(items))
		}
		var urutan []float64
		seen := map[string]bool{}
		for _, k := range ks {
			idx, ok := labelToIdx[k]
			if !ok {
				return nil, nil, fmt.Errorf("kunci %q tidak cocok dengan label opsi yang terisi", k)
			}
			if seen[k] {
				return nil, nil, fmt.Errorf("kunci berisi label %q duplikat", k)
			}
			seen[k] = true
			urutan = append(urutan, float64(idx))
		}
		return mustJSON(map[string]any{"item": items}),
			mustJSON(map[string]any{"urutan": urutan}), nil
	case "isi_ganda":
		alts := splitAndTrim(kunciStr, "|")
		if len(alts) == 0 {
			return nil, nil, fmt.Errorf("kunci wajib jawaban per kolom, dipisah |")
		}
		kolom := make([]string, 0, len(alts))
		for i := range alts {
			kolom = append(kolom, fmt.Sprintf("Kolom %d", i+1))
		}
		return mustJSON(map[string]any{"kolom": kolom}),
			mustJSON(map[string]any{"jawaban": alts}), nil
	case "teks_rumpang":
		alts := splitAndTrim(kunciStr, "|")
		if len(alts) == 0 {
			return nil, nil, fmt.Errorf("kunci wajib jawaban per rumpang, dipisah |")
		}
		return mustJSON(map[string]any{}),
			mustJSON(map[string]any{"jawaban": alts}), nil
	case "pernyataan_sikap":
		opsi := opsiFromCSV(get, rec)
		if len(opsi) < 2 {
			return nil, nil, fmt.Errorf("opsi_a..opsi_e: isi pernyataan sikap (minimal 2)")
		}
		pernyataan := make([]string, 0, len(opsi))
		for _, o := range opsi {
			pernyataan = append(pernyataan, o["teks"].(string))
		}
		skala := []float64{1, 2, 3, 4}
		return mustJSON(map[string]any{"pernyataan": pernyataan, "skala": skala}),
			mustJSON(map[string]any{"skala": skala}), nil
	}
	return nil, nil, fmt.Errorf("tipe %q belum didukung import CSV", tipe)
}
