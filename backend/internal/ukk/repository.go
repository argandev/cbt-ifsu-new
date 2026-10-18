package ukk

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ---------------------------------------------------------------------
// Model JSON
// (Dipindah dari ukk.go tanpa perubahan.)
// ---------------------------------------------------------------------

type ukkRow struct {
	ID               string     `json:"id"`
	Nama             string     `json:"nama"`
	MapelID          string     `json:"mapel_id"`
	KelasID          string     `json:"kelas_id"`
	JurusanID        string     `json:"jurusan_id"`
	Tanggal          *time.Time `json:"tanggal"`
	PengujiInternal  *string    `json:"penguji_internal"`
	PengujiEksternal *string    `json:"penguji_eksternal"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	MapelNama        *string    `json:"mapel_nama,omitempty"`
	KelasNama        *string    `json:"kelas_nama,omitempty"`
	JurusanKode      *string    `json:"jurusan_kode,omitempty"`
	JurusanNama      *string    `json:"jurusan_nama,omitempty"`
	JumlahPeserta    *int64     `json:"jumlah_peserta,omitempty"`
}

type nilaiRow struct {
	ID           string    `json:"id"`
	UkkID        string    `json:"ukk_id"`
	SiswaID      string    `json:"siswa_id"`
	NilaiTeori   *float64  `json:"nilai_teori"`
	NilaiProses  *float64  `json:"nilai_proses"`
	NilaiProduk  *float64  `json:"nilai_produk"`
	NilaiSikap   *float64  `json:"nilai_sikap"`
	NilaiPraktik *float64  `json:"nilai_praktik"`
	NilaiAkhir   *float64  `json:"nilai_akhir"`
	Lulus        *bool     `json:"lulus"`
	Predikat     *string   `json:"predikat"`
	Catatan      *string   `json:"catatan"`
	DinilaiOleh  *string   `json:"dinilai_oleh"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	SiswaNISN    *string   `json:"siswa_nisn,omitempty"`
	SiswaNama    *string   `json:"siswa_nama,omitempty"`
}

type berkasRow struct {
	ID           string    `json:"id"`
	UkkID        string    `json:"ukk_id"`
	SiswaID      string    `json:"siswa_id"`
	NamaFile     string    `json:"nama_file"`
	Ukuran       int       `json:"ukuran"`
	DiuploadOleh *string   `json:"diupload_oleh"`
	CreatedAt    time.Time `json:"created_at"`
	SiswaNISN    *string   `json:"siswa_nisn,omitempty"`
	SiswaNama    *string   `json:"siswa_nama,omitempty"`
}

type sertifikatRow struct {
	ID               string     `json:"id"`
	UkkID            string     `json:"ukk_id"`
	SiswaID          string     `json:"siswa_id"`
	NomorSertifikat  string     `json:"nomor_sertifikat"`
	DiterbitkanAt    time.Time  `json:"diterbitkan_at"`
	SiswaNISN        *string    `json:"siswa_nisn,omitempty"`
	SiswaNama        *string    `json:"siswa_nama,omitempty"`
	UkkNama          *string    `json:"ukk_nama,omitempty"`
	UkkTanggal       *time.Time `json:"ukk_tanggal,omitempty"`
	PengujiInternal  *string    `json:"penguji_internal,omitempty"`
	PengujiEksternal *string    `json:"penguji_eksternal,omitempty"`
	JurusanKode      *string    `json:"jurusan_kode,omitempty"`
	JurusanNama      *string    `json:"jurusan_nama,omitempty"`
	NilaiTeori       *float64   `json:"nilai_teori,omitempty"`
	NilaiPraktik     *float64   `json:"nilai_praktik,omitempty"`
	NilaiAkhir       *float64   `json:"nilai_akhir,omitempty"`
	Predikat         *string    `json:"predikat,omitempty"`
}

// pesertaRow adalah SATU-SATUNYA data siswa yang dibutuhkan halaman
// nilai UKK (dropdown pilih siswa). Dibatasi ke kelas UKK ini saja dan
// HANYA field id/nama/nisn/kelas_id — bukan akses master umum: tanpa
// filter global, tanpa NIS, tanpa user_id, tanpa tahun_masuk.
type pesertaRow struct {
	ID      string `json:"id"`
	Nama    string `json:"nama"`
	NISN    string `json:"nisn"`
	KelasID string `json:"kelas_id"`
}

// ---------------------------------------------------------------------
// Repository
// ---------------------------------------------------------------------

// Repository adalah lapisan akses data paket ukk (semua SQL ada di sini).
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetUkkRow membaca satu baris ukk_ujian (+ nama mapel/kelas/jurusan).
// (Dulu: getUkkRow.)
func (r *Repository) GetUkkRow(ctx context.Context, id string) (*ukkRow, error) {
	var u ukkRow
	err := r.pool.QueryRow(ctx,
		`SELECT u.id, u.nama, u.mapel_id, u.kelas_id, u.jurusan_id, u.tanggal,
		        u.penguji_internal, u.penguji_eksternal, u.status, u.created_at,
		        m.nama, k.nama, j.kode, j.nama
		 FROM ukk_ujian u
		 JOIN mapel m ON m.id = u.mapel_id
		 JOIN kelas k ON k.id = u.kelas_id
		 JOIN jurusan j ON j.id = u.jurusan_id
		 WHERE u.id = $1`, id).Scan(
		&u.ID, &u.Nama, &u.MapelID, &u.KelasID, &u.JurusanID, &u.Tanggal,
		&u.PengujiInternal, &u.PengujiEksternal, &u.Status, &u.CreatedAt,
		&u.MapelNama, &u.KelasNama, &u.JurusanKode, &u.JurusanNama)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ExistsID melaporkan apakah id ada di tabel (table dari konstanta
// internal, bukan input user). (Dulu: existsID.)
func (r *Repository) ExistsID(ctx context.Context, table, id string) bool {
	var n int
	// table berasal dari konstanta internal, bukan input user.
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM `+table+` WHERE id = $1`, id).Scan(&n)
	return n > 0
}

// CountUkk menghitung total ukk_ujian sesuai filter.
func (r *Repository) CountUkk(ctx context.Context, jurusanID, kelasID, status, search any) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM ukk_ujian u
		 WHERE ($1::uuid IS NULL OR u.jurusan_id = $1)
		   AND ($2::uuid IS NULL OR u.kelas_id = $2)
		   AND ($3::text IS NULL OR u.status = $3)
		   AND ($4::text IS NULL OR u.nama ILIKE '%' || $4 || '%')`,
		jurusanID, kelasID, status, search).Scan(&total)
	return total, err
}

// ListUkkRows membaca halaman ukk_ujian sesuai filter.
func (r *Repository) ListUkkRows(ctx context.Context, jurusanID, kelasID, status, search any, limit, offset int) ([]ukkRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.nama, u.mapel_id, u.kelas_id, u.jurusan_id, u.tanggal,
		        u.penguji_internal, u.penguji_eksternal, u.status, u.created_at,
		        m.nama, k.nama, j.kode, j.nama,
		        (SELECT COUNT(*) FROM ukk_nilai n WHERE n.ukk_id = u.id)
		 FROM ukk_ujian u
		 JOIN mapel m ON m.id = u.mapel_id
		 JOIN kelas k ON k.id = u.kelas_id
		 JOIN jurusan j ON j.id = u.jurusan_id
		 WHERE ($1::uuid IS NULL OR u.jurusan_id = $1)
		   AND ($2::uuid IS NULL OR u.kelas_id = $2)
		   AND ($3::text IS NULL OR u.status = $3)
		   AND ($4::text IS NULL OR u.nama ILIKE '%' || $4 || '%')
		 ORDER BY u.created_at DESC
		 LIMIT $5 OFFSET $6`,
		jurusanID, kelasID, status, search, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ukkRow{}
	for rows.Next() {
		var u ukkRow
		var jumlah int64
		if err := rows.Scan(&u.ID, &u.Nama, &u.MapelID, &u.KelasID, &u.JurusanID,
			&u.Tanggal, &u.PengujiInternal, &u.PengujiEksternal, &u.Status, &u.CreatedAt,
			&u.MapelNama, &u.KelasNama, &u.JurusanKode, &u.JurusanNama, &jumlah); err != nil {
			return nil, err
		}
		u.JumlahPeserta = &jumlah
		items = append(items, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// InsertUkk menyisipkan ukk_ujian baru; mengembalikan id.
func (r *Repository) InsertUkk(ctx context.Context, nama, mapelID, kelasID, jurusanID string, tanggalArg, pengujiInternal, pengujiEksternal any, status string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO ukk_ujian (nama, mapel_id, kelas_id, jurusan_id, tanggal, penguji_internal, penguji_eksternal, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		nama, mapelID, kelasID, jurusanID,
		tanggalArg, pengujiInternal, pengujiEksternal, status).Scan(&id)
	return id, err
}

// UpdateUkk mengubah satu baris ukk_ujian.
func (r *Repository) UpdateUkk(ctx context.Context, id, nama, mapelID, kelasID, jurusanID string, tanggalArg, pengujiInternal, pengujiEksternal any, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE ukk_ujian SET nama = $2, mapel_id = $3, kelas_id = $4, jurusan_id = $5,
		        tanggal = $6, penguji_internal = $7, penguji_eksternal = $8, status = $9
		 WHERE id = $1`,
		id, nama, mapelID, kelasID, jurusanID,
		tanggalArg, pengujiInternal, pengujiEksternal, status)
	return err
}

// DeleteUkk menghapus satu baris ukk_ujian.
func (r *Repository) DeleteUkk(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM ukk_ujian WHERE id = $1`, id)
	return err
}

// CountAnak menghitung baris nilai/berkas/sertifikat satu UKK
// (gerbang hapus). Error tiap query diabaikan seperti kode lama.
func (r *Repository) CountAnak(ctx context.Context, id string) (nNilai, nBerkas, nSert int) {
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM ukk_nilai WHERE ukk_id = $1`, id).Scan(&nNilai)
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM ukk_berkas WHERE ukk_id = $1`, id).Scan(&nBerkas)
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM ukk_sertifikat WHERE ukk_id = $1`, id).Scan(&nSert)
	return nNilai, nBerkas, nSert
}

// ListNilaiRows membaca semua nilai satu UKK.
func (r *Repository) ListNilaiRows(ctx context.Context, ukkID string) ([]nilaiRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT n.id, n.ukk_id, n.siswa_id, n.nilai_teori, n.nilai_proses,
		        n.nilai_produk, n.nilai_sikap, n.nilai_praktik, n.nilai_akhir,
		        n.lulus, n.predikat, n.catatan, n.dinilai_oleh,
		        n.created_at, n.updated_at, s.nisn, s.nama
		 FROM ukk_nilai n JOIN siswa s ON s.id = n.siswa_id
		 WHERE n.ukk_id = $1 ORDER BY s.nama ASC`, ukkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []nilaiRow{}
	for rows.Next() {
		var n nilaiRow
		if err := rows.Scan(&n.ID, &n.UkkID, &n.SiswaID, &n.NilaiTeori, &n.NilaiProses,
			&n.NilaiProduk, &n.NilaiSikap, &n.NilaiPraktik, &n.NilaiAkhir,
			&n.Lulus, &n.Predikat, &n.Catatan, &n.DinilaiOleh,
			&n.CreatedAt, &n.UpdatedAt, &n.SiswaNISN, &n.SiswaNama); err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// ListPesertaRows membaca siswa aktif satu kelas (dropdown nilai UKK).
func (r *Repository) ListPesertaRows(ctx context.Context, kelasID string) ([]pesertaRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, nama, nisn, kelas_id FROM siswa
		 WHERE kelas_id = $1 AND is_active ORDER BY nama ASC`, kelasID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []pesertaRow{}
	for rows.Next() {
		var p pesertaRow
		if err := rows.Scan(&p.ID, &p.Nama, &p.NISN, &p.KelasID); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// SiswaIDByNilaiID membaca siswa_id dari satu baris nilai.
// Tidak ada -> error (pemanggil memetakan ke 404).
func (r *Repository) SiswaIDByNilaiID(ctx context.Context, nilaiID, ukkID string) (string, error) {
	var siswaID string
	err := r.pool.QueryRow(ctx,
		`SELECT siswa_id FROM ukk_nilai WHERE id = $1 AND ukk_id = $2`,
		nilaiID, ukkID).Scan(&siswaID)
	return siswaID, err
}

// UpdateNilaiRow memperbarui satu baris nilai (jalur PUT per nilai id).
func (r *Repository) UpdateNilaiRow(ctx context.Context, nilaiID string, in nilaiInput, praktik, akhir *float64, lulus *bool, predikat *string, userID string) (*nilaiRow, error) {
	var row nilaiRow
	err := r.pool.QueryRow(ctx,
		`UPDATE ukk_nilai SET nilai_teori = $2, nilai_proses = $3, nilai_produk = $4,
		        nilai_sikap = $5, nilai_praktik = $6, nilai_akhir = $7, lulus = $8,
		        predikat = $9, catatan = $10, dinilai_oleh = $11, updated_at = now()
		 WHERE id = $1 RETURNING id, ukk_id, siswa_id, nilai_teori, nilai_proses,
		        nilai_produk, nilai_sikap, nilai_praktik, nilai_akhir, lulus,
		        predikat, catatan, dinilai_oleh, created_at, updated_at`,
		nilaiID, in.NilaiTeori, in.NilaiProses, in.NilaiProduk, in.NilaiSikap,
		praktik, akhir, lulus, predikat, strPtrOrNil(in.Catatan), userID).Scan(
		&row.ID, &row.UkkID, &row.SiswaID, &row.NilaiTeori, &row.NilaiProses,
		&row.NilaiProduk, &row.NilaiSikap, &row.NilaiPraktik, &row.NilaiAkhir,
		&row.Lulus, &row.Predikat, &row.Catatan, &row.DinilaiOleh,
		&row.CreatedAt, &row.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// NilaiExists melaporkan apakah sudah ada nilai untuk (ukk_id, siswa_id).
func (r *Repository) NilaiExists(ctx context.Context, ukkID, siswaID string) bool {
	var existed bool
	_ = r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM ukk_nilai WHERE ukk_id = $1 AND siswa_id = $2)`,
		ukkID, siswaID).Scan(&existed)
	return existed
}

// UpsertNilaiRow menyisipkan/memperbarui nilai per (ukk_id, siswa_id)
// (jalur POST).
func (r *Repository) UpsertNilaiRow(ctx context.Context, ukkID, siswaID string, in nilaiInput, praktik, akhir *float64, lulus *bool, predikat *string, userID string) (*nilaiRow, error) {
	var row nilaiRow
	err := r.pool.QueryRow(ctx,
		`INSERT INTO ukk_nilai (ukk_id, siswa_id, nilai_teori, nilai_proses, nilai_produk,
		        nilai_sikap, nilai_praktik, nilai_akhir, lulus, predikat, catatan, dinilai_oleh)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 ON CONFLICT (ukk_id, siswa_id) DO UPDATE
		 SET nilai_teori = EXCLUDED.nilai_teori, nilai_proses = EXCLUDED.nilai_proses,
		     nilai_produk = EXCLUDED.nilai_produk, nilai_sikap = EXCLUDED.nilai_sikap,
		     nilai_praktik = EXCLUDED.nilai_praktik, nilai_akhir = EXCLUDED.nilai_akhir,
		     lulus = EXCLUDED.lulus, predikat = EXCLUDED.predikat,
		     catatan = EXCLUDED.catatan, dinilai_oleh = EXCLUDED.dinilai_oleh,
		     updated_at = now()
		 RETURNING id, ukk_id, siswa_id, nilai_teori, nilai_proses,
		        nilai_produk, nilai_sikap, nilai_praktik, nilai_akhir, lulus,
		        predikat, catatan, dinilai_oleh, created_at, updated_at`,
		ukkID, siswaID, in.NilaiTeori, in.NilaiProses, in.NilaiProduk, in.NilaiSikap,
		praktik, akhir, lulus, predikat, strPtrOrNil(in.Catatan), userID).Scan(
		&row.ID, &row.UkkID, &row.SiswaID, &row.NilaiTeori, &row.NilaiProses,
		&row.NilaiProduk, &row.NilaiSikap, &row.NilaiPraktik, &row.NilaiAkhir,
		&row.Lulus, &row.Predikat, &row.Catatan, &row.DinilaiOleh,
		&row.CreatedAt, &row.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// SiswaNISNNama membaca nisn + nama siswa (untuk melengkapi respons).
// Error diabaikan seperti kode lama.
func (r *Repository) SiswaNISNNama(ctx context.Context, siswaID string) (*string, *string) {
	var nisn, nama *string
	_ = r.pool.QueryRow(ctx, `SELECT nisn, nama FROM siswa WHERE id = $1`, siswaID).
		Scan(&nisn, &nama)
	return nisn, nama
}

// ResolveSiswaID memetakan user login ke siswa.id via siswa.user_id.
// Mengembalikan ("", false) bila tidak ada profil siswa untuk user ini.
// (Dulu: resolveSiswaID.)
func (r *Repository) ResolveSiswaID(ctx context.Context, userID string) (string, bool) {
	var id string
	err := r.pool.QueryRow(ctx, `SELECT id FROM siswa WHERE user_id = $1`, userID).Scan(&id)
	if err != nil {
		return "", false
	}
	return id, true
}

// ListBerkasRows membaca berkas satu UKK, opsional filter siswa.
func (r *Repository) ListBerkasRows(ctx context.Context, ukkID string, filterSiswa *string) ([]berkasRow, error) {
	var filter any
	if filterSiswa != nil {
		filter = *filterSiswa
	}
	rows, err := r.pool.Query(ctx,
		`SELECT b.id, b.ukk_id, b.siswa_id, b.nama_file, b.ukuran,
		        b.diupload_oleh, b.created_at, s.nisn, s.nama
		 FROM ukk_berkas b JOIN siswa s ON s.id = b.siswa_id
		 WHERE b.ukk_id = $1 AND ($2::uuid IS NULL OR b.siswa_id = $2)
		 ORDER BY b.created_at DESC`, ukkID, filter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []berkasRow{}
	for rows.Next() {
		var b berkasRow
		if err := rows.Scan(&b.ID, &b.UkkID, &b.SiswaID, &b.NamaFile, &b.Ukuran,
			&b.DiuploadOleh, &b.CreatedAt, &b.SiswaNISN, &b.SiswaNama); err != nil {
			return nil, err
		}
		items = append(items, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// InsertBerkas mencatat satu berkas praktik.
func (r *Repository) InsertBerkas(ctx context.Context, ukkID, siswaID, namaFile, relPath string, ukuran int64, userID string) (*berkasRow, error) {
	var b berkasRow
	err := r.pool.QueryRow(ctx,
		`INSERT INTO ukk_berkas (ukk_id, siswa_id, nama_file, path, ukuran, diupload_oleh)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, ukk_id, siswa_id, nama_file, ukuran, diupload_oleh, created_at`,
		ukkID, siswaID, namaFile, relPath, ukuran, userID).Scan(
		&b.ID, &b.UkkID, &b.SiswaID, &b.NamaFile, &b.Ukuran, &b.DiuploadOleh, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// GetBerkasRow membaca satu berkas (+ path tersimpan).
// Tidak ada -> error (pemanggil memetakan ke 404).
func (r *Repository) GetBerkasRow(ctx context.Context, berkasID, ukkID string) (*berkasRow, string, error) {
	var b berkasRow
	var storedPath string
	err := r.pool.QueryRow(ctx,
		`SELECT id, ukk_id, siswa_id, nama_file, ukuran, diupload_oleh, created_at, path
		 FROM ukk_berkas WHERE id = $1 AND ukk_id = $2`,
		berkasID, ukkID).Scan(&b.ID, &b.UkkID, &b.SiswaID, &b.NamaFile, &b.Ukuran,
		&b.DiuploadOleh, &b.CreatedAt, &storedPath)
	if err != nil {
		return nil, "", err
	}
	return &b, storedPath, nil
}

// GetSertifikatRow membaca satu sertifikat (+ info UKK & nilai).
// Tidak ada -> error (pemanggil memetakan ke 404).
func (r *Repository) GetSertifikatRow(ctx context.Context, ukkID, siswaID string) (*sertifikatRow, error) {
	var s sertifikatRow
	err := r.pool.QueryRow(ctx,
		`SELECT s.id, s.ukk_id, s.siswa_id, s.nomor_sertifikat, s.diterbitkan_at,
		        si.nisn, si.nama, u.nama, u.tanggal, u.penguji_internal, u.penguji_eksternal,
		        j.kode, j.nama, n.nilai_teori, n.nilai_praktik, n.nilai_akhir, n.predikat
		 FROM ukk_sertifikat s
		 JOIN siswa si ON si.id = s.siswa_id
		 JOIN ukk_ujian u ON u.id = s.ukk_id
		 JOIN jurusan j ON j.id = u.jurusan_id
		 LEFT JOIN ukk_nilai n ON n.ukk_id = s.ukk_id AND n.siswa_id = s.siswa_id
		 WHERE s.ukk_id = $1 AND s.siswa_id = $2`,
		ukkID, siswaID).Scan(
		&s.ID, &s.UkkID, &s.SiswaID, &s.NomorSertifikat, &s.DiterbitkanAt,
		&s.SiswaNISN, &s.SiswaNama, &s.UkkNama, &s.UkkTanggal, &s.PengujiInternal,
		&s.PengujiEksternal, &s.JurusanKode, &s.JurusanNama,
		&s.NilaiTeori, &s.NilaiPraktik, &s.NilaiAkhir, &s.Predikat)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetLulus membaca flag lulus satu siswa pada satu UKK.
// Tidak ada baris -> (nil, nil); error lain dikembalikan.
func (r *Repository) GetLulus(ctx context.Context, ukkID, siswaID string) (*bool, error) {
	var lulus *bool
	err := r.pool.QueryRow(ctx,
		`SELECT lulus FROM ukk_nilai WHERE ukk_id = $1 AND siswa_id = $2`,
		ukkID, siswaID).Scan(&lulus)
	if err != nil {
		return nil, err
	}
	return lulus, nil
}

// CountSertifikatUrut menghitung sertifikat per jurusan + tahun
// (untuk nomor urut penerbitan).
func (r *Repository) CountSertifikatUrut(ctx context.Context, jurusanID string, tahun int) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM ukk_sertifikat s JOIN ukk_ujian u ON u.id = s.ukk_id
		 WHERE u.jurusan_id = $1 AND EXTRACT(YEAR FROM s.diterbitkan_at) = $2`,
		jurusanID, tahun).Scan(&count)
	return count, err
}

// InsertSertifikat menerbitkan satu sertifikat.
func (r *Repository) InsertSertifikat(ctx context.Context, ukkID, siswaID, nomor string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO ukk_sertifikat (ukk_id, siswa_id, nomor_sertifikat) VALUES ($1, $2, $3)`,
		ukkID, siswaID, nomor)
	return err
}
