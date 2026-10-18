// Package rekap — rekap nilai, berita acara, export CSV, finalisasi, remedial.
//
// Asumsi skema tim lain (migrasi 0001-0004):
//
//	jadwal_ujian(id, paket_id, kelas_id, ...), mapel(id, nama),
//	siswa(id, nisn, nama, kelas_id), kelas(id, nama),
//	attempt(id, jadwal_id, siswa_id, status),
//	jawaban(id, attempt_id, ...), settings(key, value jsonb).
//
// Status attempt yang dihitung "hadir" di berita acara: 'selesai', 'diblokir'.
// Predikat (ANALISIS Bab 8 / UKK): <70 Belum Tuntas, 70-79 Cukup,
// 80-90 Baik, 91-100 Sangat Baik. Ambang lulus rekap: nilai >= 70.
package rekap

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// Predikat menghitung predikat dari nilai 0-100.
func Predikat(nilai float64) string {
	switch {
	case nilai < 70:
		return "Belum Tuntas"
	case nilai < 80:
		return "Cukup"
	case nilai <= 90:
		return "Baik"
	default:
		return "Sangat Baik"
	}
}

type nilaiRow struct {
	ID            string  `json:"id"`
	AttemptID     string  `json:"attempt_id"`
	JadwalID      string  `json:"jadwal_id"`
	SiswaID       string  `json:"siswa_id"`
	Nilai         float64 `json:"nilai"`
	Status        string  `json:"status"`
	Predikat      *string `json:"predikat"`
	NISN          string  `json:"nisn"`
	NamaSiswa     string  `json:"nama"`
	NamaKelas     string  `json:"kelas"`
	JumlahDijawab int     `json:"jumlah_dijawab"`
}

func writeError(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

func writeOK(c *echo.Context, data any) error {
	return c.JSON(http.StatusOK, map[string]any{"data": data})
}

// tampilHasilKeSiswa membaca setting tampil_hasil_ke_siswa (default false).
func tampilHasilKeSiswa(c *echo.Context, pool *pgxpool.Pool) bool {
	var raw json.RawMessage
	err := pool.QueryRow(c.Request().Context(),
		`SELECT value FROM settings WHERE key = 'tampil_hasil_ke_siswa'`).Scan(&raw)
	if err != nil {
		return false
	}
	var v bool
	if err := json.Unmarshal(raw, &v); err != nil {
		return false
	}
	return v
}

const listNilaiSQL = `
	SELECT na.id, na.attempt_id, na.jadwal_id, na.siswa_id, na.nilai, na.status, na.predikat,
	       s.nisn, s.nama, COALESCE(k.nama, ''), 
	       -- D3: hanya jawaban yang benar-benar diisi (jawaban IS NOT NULL).
	       -- Baris jawaban dengan jawaban NULL dibuat saat siswa menandai
	       -- ragu-ragu saja — bukan jawaban — dan tidak boleh terhitung
	       -- sebagai "Dijawab" (samakan dengan Monitor).
	       (SELECT COUNT(*) FROM jawaban j WHERE j.attempt_id = na.attempt_id AND j.jawaban IS NOT NULL)
	FROM nilai_akhir na
	JOIN siswa s ON s.id = na.siswa_id
	LEFT JOIN kelas k ON k.id = s.kelas_id
	WHERE na.jadwal_id = $1`

func queryNilai(c *echo.Context, pool *pgxpool.Pool, jadwalID string, siswaID *string, onlyFinal bool) ([]nilaiRow, error) {
	ctx := c.Request().Context()
	query := listNilaiSQL
	args := []any{jadwalID}
	if siswaID != nil {
		query += ` AND na.siswa_id = $2`
		args = append(args, *siswaID)
	}
	if onlyFinal {
		query += ` AND na.status = 'final'`
	}
	query += ` ORDER BY s.nama ASC`
	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []nilaiRow{}
	for rows.Next() {
		var r nilaiRow
		var predikat *string
		if err := rows.Scan(&r.ID, &r.AttemptID, &r.JadwalID, &r.SiswaID, &r.Nilai, &r.Status, &predikat,
			&r.NISN, &r.NamaSiswa, &r.NamaKelas, &r.JumlahDijawab); err != nil {
			return nil, err
		}
		r.Predikat = predikat
		out = append(out, r)
	}
	return out, rows.Err()
}

func statistik(rows []nilaiRow) map[string]any {
	if len(rows) == 0 {
		return map[string]any{
			"jumlah_peserta": 0, "rata_rata": nil, "tertinggi": nil, "terendah": nil,
			"jumlah_lulus": 0, "jumlah_menunggu_koreksi": 0,
		}
	}
	sum := 0.0
	max := rows[0].Nilai
	min := rows[0].Nilai
	lulus := 0
	menunggu := 0
	for _, r := range rows {
		sum += r.Nilai
		max = math.Max(max, r.Nilai)
		min = math.Min(min, r.Nilai)
		if r.Nilai >= 70 {
			lulus++
		}
		if r.Status == "menunggu_koreksi" {
			menunggu++
		}
	}
	avg := math.Round(sum/float64(len(rows))*100) / 100
	return map[string]any{
		"jumlah_peserta": len(rows), "rata_rata": avg, "tertinggi": max, "terendah": min,
		"jumlah_lulus": lulus, "jumlah_menunggu_koreksi": menunggu,
	}
}

// Register mendaftarkan endpoint rekap.
//
//	GET  /api/v1/rekap?jadwal_id=
//	GET  /api/v1/rekap/:jadwalId/export.csv
//	GET  /api/v1/rekap/:jadwalId/berita-acara
//	POST /api/v1/rekap/:jadwalId/finalisasi
//	POST /api/v1/rekap/:jadwalId/remedial
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	// GET /api/v1/rekap?jadwal_id=
	e.GET("/api/v1/rekap", func(c *echo.Context) error {
		jadwalID := strings.TrimSpace(c.QueryParam("jadwal_id"))
		if jadwalID == "" {
			return writeError(c, http.StatusBadRequest, "validation_error", "Parameter jadwal_id wajib diisi.")
		}
		userID, role, _, siswaID, ok := audit.CurrentUser(c)
		_ = userID
		if !ok {
			return writeError(c, http.StatusUnauthorized, "unauthorized", "Harus login.")
		}
		if role == "siswa" {
			// Siswa hanya melihat nilainya sendiri, hanya jika final + tampil_hasil.
			if !tampilHasilKeSiswa(c, pool) {
				return writeOK(c, map[string]any{"items": []nilaiRow{}, "statistik": statistik(nil)})
			}
			rows, err := queryNilai(c, pool, jadwalID, &siswaID, true)
			if err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca rekap.")
			}
			return writeOK(c, map[string]any{"items": rows, "statistik": statistik(rows)})
		}
		if !audit.HasPermission(c, "rekap.read") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat rekap.")
		}
		rows, err := queryNilai(c, pool, jadwalID, nil, false)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca rekap.")
		}
		return writeOK(c, map[string]any{"items": rows, "statistik": statistik(rows)})
	})

	// GET /api/v1/rekap/:jadwalId/export.csv
	e.GET("/api/v1/rekap/:jadwalId/export.csv", func(c *echo.Context) error {
		jadwalID := c.Param("jadwalId")
		if !audit.HasPermission(c, "rekap.export") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin export rekap.")
		}
		rows, err := queryNilai(c, pool, jadwalID, nil, false)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca rekap.")
		}
		userID, _, _, _, _ := audit.CurrentUser(c)
		_ = audit.Catat(c.Request().Context(), pool, userID, "rekap.export", "rekap", jadwalID,
			map[string]any{"jadwal_id": jadwalID, "jumlah_baris": len(rows)}, c.RealIP())

		filename := fmt.Sprintf("rekap-%s-%s.csv", jadwalID, time.Now().Format("2006-01-02"))
		c.Response().Header().Set("Content-Type", "text/csv; charset=utf-8")
		c.Response().Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
		w := csv.NewWriter(c.Response())
		// BOM UTF-8 agar Excel membuka file dengan benar.
		_, _ = c.Response().Write([]byte{0xEF, 0xBB, 0xBF})
		_ = w.Write([]string{"nisn", "nama", "kelas", "nilai", "predikat", "status"})
		for _, r := range rows {
			pred := ""
			if r.Predikat != nil {
				pred = *r.Predikat
			}
			_ = w.Write([]string{r.NISN, r.NamaSiswa, r.NamaKelas,
				strconv.FormatFloat(r.Nilai, 'f', 2, 64), pred, r.Status})
		}
		w.Flush()
		return nil
	})

	// GET /api/v1/rekap/:jadwalId/berita-acara
	e.GET("/api/v1/rekap/:jadwalId/berita-acara", func(c *echo.Context) error {
		jadwalID := c.Param("jadwalId")
		if !audit.HasPermission(c, "rekap.read") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat berita acara.")
		}
		ctx := c.Request().Context()
		var jadwalInfo struct {
			ID         string
			MapelNama  string
			KelasNama  string
			Tanggal    any
			JamMulai   any
			JamSelesai any
		}
		err := pool.QueryRow(ctx,
			// Tanggal/jam diambil sebagai TEXT: pgtype.Time/Date yang di-scan
			// ke `any` terserialisasi jadi objek {"Microseconds":...,"Valid":...}
			// sehingga frontend crash ("Objects are not valid as a React child")
			// saat merender jam_mulai/jam_selesai di BeritaAcaraPage.
			`SELECT j.id, COALESCE(m.nama, ''), COALESCE(k.nama, ''), j.tanggal::text, j.jam_mulai::text, j.jam_selesai::text
			 FROM jadwal_ujian j
			 JOIN paket_ujian p ON p.id = j.paket_id
			 LEFT JOIN mapel m ON m.id = p.mapel_id
			 LEFT JOIN kelas k ON k.id = j.kelas_id
			 WHERE j.id = $1`, jadwalID).
			Scan(&jadwalInfo.ID, &jadwalInfo.MapelNama, &jadwalInfo.KelasNama,
				&jadwalInfo.Tanggal, &jadwalInfo.JamMulai, &jadwalInfo.JamSelesai)
		if err != nil {
			// Fallback ke info minimal (jadwal tanpa mapel).
			err2 := pool.QueryRow(ctx,
				`SELECT j.id, COALESCE(m.nama, '') FROM jadwal_ujian j
				 JOIN paket_ujian p ON p.id = j.paket_id
				 LEFT JOIN mapel m ON m.id = p.mapel_id WHERE j.id = $1`, jadwalID).
				Scan(&jadwalInfo.ID, &jadwalInfo.MapelNama)
			if err2 != nil {
				return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
			}
		}
		rows, err := queryNilai(c, pool, jadwalID, nil, false)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca rekap.")
		}
		var jumlahAttempt, jumlahHadir int
		err = pool.QueryRow(ctx,
			`SELECT COUNT(*), COUNT(*) FILTER (WHERE status IN ('selesai','diblokir'))
			 FROM attempt WHERE jadwal_id = $1`, jadwalID).Scan(&jumlahAttempt, &jumlahHadir)
		if err != nil {
			jumlahAttempt = len(rows)
			jumlahHadir = len(rows)
		}
		stats := statistik(rows)
		lulus, _ := stats["jumlah_lulus"].(int)
		// Kelas: dari jadwal bila ada, bila kosong gabungkan kelas peserta.
		kelas := jadwalInfo.KelasNama
		if kelas == "" {
			seen := map[string]bool{}
			names := []string{}
			for _, r := range rows {
				if r.NamaKelas != "" && !seen[r.NamaKelas] {
					seen[r.NamaKelas] = true
					names = append(names, r.NamaKelas)
				}
			}
			kelas = strings.Join(names, ", ")
		}
		return writeOK(c, map[string]any{
			"jadwal_id":      jadwalID,
			"mapel":          jadwalInfo.MapelNama,
			"kelas":          kelas,
			"tanggal":        jadwalInfo.Tanggal,
			"jam_mulai":      jadwalInfo.JamMulai,
			"jam_selesai":    jadwalInfo.JamSelesai,
			"jumlah_peserta": jumlahAttempt,
			"hadir":          jumlahHadir,
			"tidak_hadir":    jumlahAttempt - jumlahHadir,
			"rata_rata":      stats["rata_rata"],
			"lulus":          lulus,
			"catatan":        "",
			"generated_at":   time.Now().Format(time.RFC3339),
		})
	})

	// POST /api/v1/rekap/:jadwalId/finalisasi
	e.POST("/api/v1/rekap/:jadwalId/finalisasi", func(c *echo.Context) error {
		jadwalID := c.Param("jadwalId")
		if !audit.HasPermission(c, "rekap.finalisasi") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin finalisasi nilai.")
		}
		ctx := c.Request().Context()
		tx, err := pool.Begin(ctx)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memulai finalisasi.")
		}
		defer tx.Rollback(ctx)
		rows, err := tx.Query(ctx,
			`UPDATE nilai_akhir SET status = 'final',
				predikat = CASE WHEN nilai < 70 THEN 'Belum Tuntas'
					WHEN nilai < 80 THEN 'Cukup'
					WHEN nilai <= 90 THEN 'Baik'
					ELSE 'Sangat Baik' END,
				updated_at = now()
			 WHERE jadwal_id = $1 RETURNING id, nilai`, jadwalID)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal finalisasi nilai.")
		}
		count := 0
		for rows.Next() {
			var id string
			var nilai float64
			if err := rows.Scan(&id, &nilai); err != nil {
				rows.Close()
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal finalisasi nilai.")
			}
			count++
		}
		rows.Close()
		if err := tx.Commit(ctx); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal finalisasi nilai.")
		}
		userID, _, _, _, _ := audit.CurrentUser(c)
		_ = audit.Catat(ctx, pool, userID, "rekap.finalisasi", "rekap", jadwalID,
			map[string]any{"jadwal_id": jadwalID, "jumlah_difinalisasi": count}, c.RealIP())
		return writeOK(c, map[string]any{"jadwal_id": jadwalID, "jumlah_difinalisasi": count, "status": "final"})
	})

	// POST /api/v1/rekap/:jadwalId/remedial {siswa_id, nilai_baru}
	e.POST("/api/v1/rekap/:jadwalId/remedial", func(c *echo.Context) error {
		jadwalID := c.Param("jadwalId")
		if !audit.HasPermission(c, "rekap.remedial") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin remedial.")
		}
		var body struct {
			SiswaID   string  `json:"siswa_id"`
			NilaiBaru float64 `json:"nilai_baru"`
		}
		if err := c.Bind(&body); err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
		}
		if strings.TrimSpace(body.SiswaID) == "" {
			return writeError(c, http.StatusBadRequest, "validation_error", "siswa_id wajib diisi.")
		}
		if body.NilaiBaru < 0 || body.NilaiBaru > 100 {
			return writeError(c, http.StatusBadRequest, "validation_error", "nilai_baru harus 0-100.")
		}
		ctx := c.Request().Context()
		tx, err := pool.Begin(ctx)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memulai remedial.")
		}
		defer tx.Rollback(ctx)
		if _, err := tx.Exec(ctx,
			`INSERT INTO remedial (jadwal_id, siswa_id, nilai) VALUES ($1, $2, $3)`,
			jadwalID, body.SiswaID, body.NilaiBaru); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyimpan remedial.")
		}
		var nilaiAwal float64
		var nilaiAkhirID string
		err = tx.QueryRow(ctx,
			`SELECT id, nilai FROM nilai_akhir WHERE jadwal_id = $1 AND siswa_id = $2`,
			jadwalID, body.SiswaID).Scan(&nilaiAkhirID, &nilaiAwal)
		if err != nil {
			return writeError(c, http.StatusNotFound, "not_found", "Nilai akhir siswa pada jadwal ini tidak ditemukan.")
		}
		nilaiAkhir := math.Max(nilaiAwal, body.NilaiBaru)
		if _, err := tx.Exec(ctx,
			`UPDATE nilai_akhir SET nilai = $1, predikat = $2, updated_at = now() WHERE id = $3`,
			nilaiAkhir, Predikat(nilaiAkhir), nilaiAkhirID); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memperbarui nilai akhir.")
		}
		if err := tx.Commit(ctx); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyimpan remedial.")
		}
		userID, _, _, _, _ := audit.CurrentUser(c)
		_ = audit.Catat(ctx, pool, userID, "rekap.remedial", "nilai_akhir", nilaiAkhirID,
			map[string]any{"jadwal_id": jadwalID, "siswa_id": body.SiswaID,
				"nilai_awal": nilaiAwal, "nilai_remedial": body.NilaiBaru, "nilai_akhir": nilaiAkhir}, c.RealIP())
		return writeOK(c, map[string]any{
			"jadwal_id": jadwalID, "siswa_id": body.SiswaID,
			"nilai_awal": nilaiAwal, "nilai_remedial": body.NilaiBaru,
			"nilai_akhir": nilaiAkhir, "predikat": Predikat(nilaiAkhir),
		})
	})
}
