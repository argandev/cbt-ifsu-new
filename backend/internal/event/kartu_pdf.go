// Kartu peserta PDF (server-side) — Tim Backend Kartu PDF.
//
// PDF dibuat di BACKEND (bukan window.print di frontend). Satu kartu =
// identitas siswa sekali + tabel SEMUA jadwal kelasnya + kotak login +
// tanda tangan + footer catatan. PDF gabungan (bulk staf) = 1 halaman
// per siswa (kartu tiap siswa tetap memuat semua jadwalnya).
//
// File ini HANYA berisi generator + handler PDF + loader data kartu PDF.
// Endpoint JSON kartu lama (legacy) tetap ada di event.go dan tidak
// diubah. Logika penomoran peserta (buildKartuList di event.go) dipakai
// ulang apa adanya — TIDAK diduplikasi di sini.
//
// Layout sengaja dipecah per bagian (kop, identitas, tabel, kotak login,
// tanda tangan, footer) dengan konstanta margin/lebar kolom di satu
// tempat, agar Tim Format mudah memoles tanpa menyentuh logika data.
package event

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jung-kurt/gofpdf"
	"github.com/labstack/echo/v5"
)

// ---------------------------------------------------------------------------
// Konstanta layout — satu tempat, mudah dipoles Tim Format.
// ---------------------------------------------------------------------------

const (
	// kartuPDFMargin: margin kiri/kanan/atas (mm). A4 portrait 210x297.
	kartuPDFMargin = 15.0
	// kartuPDFLebarUsable = 210 - 2*15 = 180 mm.
	kartuPDFLebarUsable = 180.0
	// kartuPDFBatasBawah: garis Y terakhir isi (mm) sebelum footer.
	// Footer (catatan + nomor halaman) dicetak di y = 297-12.
	kartuPDFBatasBawah = 268.0
	// kartuPDFMaksSiswaBulk: batas wajar PDF gabungan; lebih dari ini -> 400.
	kartuPDFMaksSiswaBulk = 500

	// --- Ukuran tipografi kop (mm/pt) ---------------------------------
	// Nama sekolah paling menonjol (16 tebal), judul 15 tebal, nama
	// event 11 reguler. Judul ditulis sebagai satu string utuh (TANPA
	// spasi sisipan antar huruf) agar tetap terbaca utuh oleh
	// pdftotext / pencarian teks; "spasi huruf konsisten" dijaga lewat
	// tinggi baris + jarak antar blok yang tetap.
	kartuPDFUkuranNamaSekolah = 16.0
	kartuPDFUkuranJudul       = 15.0
	kartuPDFUkuranEvent       = 11.0

	// --- Blok identitas ------------------------------------------------
	// Label lebar tetap 32 mm + ":" 5 mm, nilai mengisi sisanya.
	// Inset 2 mm dari margin agar teks tidak mepet tepi kiri.
	kartuPDFLebarLabelIdentitas = 32.0
	kartuPDFLebarTitikDua       = 5.0
	kartuPDFTinggiBarisIdent    = 7.0
	kartuPDFInsetIsi            = 2.0

	// --- Tabel jadwal ---------------------------------------------------
	kartuPDFTinggiHeaderTabel = 8.0
	kartuPDFTinggiBarisTabel  = 7.5

	// --- Kotak login ----------------------------------------------------
	// Label lebar tetap 40 mm + ":" 5 mm, nilai mengisi sisanya.
	// Tinggi total = padding atas 2.5 + judul 6 + 2 baris x 6 +
	// padding bawah 2.5 = 23 mm.
	kartuPDFLebarLabelLogin = 40.0
	kartuPDFTinggiKotakLogin = 23.0
	kartuPDFTinggiBarisLogin = 6.0

	// --- Tanda tangan ---------------------------------------------------
	// Garis tanda tangan: inset 12 mm dari tiap sisi kolom -> kedua
	// garis sama panjang (66 mm). Ruang tanda tangan 20 mm.
	kartuPDFInsetGarisTTD = 12.0
	kartuPDFRuangTTD      = 20.0
)

var (
	// kartuPDFHeaderKolom + kartuPDFLebarKolom: 7 kolom, jumlah = 180 mm.
	// Kolom Nomor Peserta sengaja paling lebar kedua: nomor TIDAK boleh
	// terpotong ("...") karena itu identitas utama peserta di ruangan.
	// Mapel/Paket/Ruangan yang sangat panjang dipotong dengan "...".
	kartuPDFHeaderKolom = []string{"No", "Nomor Peserta", "Mapel", "Paket", "Tanggal", "Jam", "Ruangan"}
	kartuPDFLebarKolom  = []float64{8, 44, 38, 26, 22, 24, 18}

	kartuPDFJudul        = "Kartu Peserta Ujian"
	kartuPDFEventDefault = "Ujian Sekolah"
	kartuPDFCatatan      = "Kartu ini wajib dibawa saat ujian. Token ujian diberikan proktor di ruangan."
)

var errKartuTerlaluBanyak = errors.New("jumlah siswa melebihi batas PDF gabungan")

// ---------------------------------------------------------------------------
// Model data kartu PDF (tidak terikat DB — bisa di-test tanpa Postgres).
// ---------------------------------------------------------------------------

// kartuPDFJadwal adalah satu baris tabel jadwal di kartu.
type kartuPDFJadwal struct {
	NomorPeserta string // nomor peserta PER JADWAL (dari buildKartuList)
	Mapel        string
	Paket        string
	Tanggal      string // sudah diformat tampil, mis. "17 Okt 2026"
	Jam          string // mis. "07:30 - 09:30"
	Ruangan      string // "-" bila belum ditugaskan
	EventNama    string // nama event jadwal ini ("" bila tanpa event)
}

// kartuPDFData adalah satu kartu lengkap untuk satu siswa.
type kartuPDFData struct {
	NamaSekolah string // dari settings; "" bila belum diatur -> baris kop dilewati
	NamaEvent   string // nama event bila semua jadwal satu event, selain itu "Ujian Sekolah"
	NamaSiswa   string
	NISN        string
	Kelas       string
	Jurusan     string
	Username    string
	Jadwal      []kartuPDFJadwal
}

// namaEventGabungan: bila semua jadwal berbagi SATU nama event yang sama,
// pakai nama itu; bila campur/tanpa event -> "Ujian Sekolah".
func namaEventGabungan(jadwal []kartuPDFJadwal) string {
	uniques := map[string]bool{}
	for _, j := range jadwal {
		if strings.TrimSpace(j.EventNama) != "" {
			uniques[strings.TrimSpace(j.EventNama)] = true
		}
	}
	if len(uniques) == 1 {
		for nama := range uniques {
			return nama
		}
	}
	return kartuPDFEventDefault
}

// ---------------------------------------------------------------------------
// Helper teks / tanggal.
// ---------------------------------------------------------------------------

// potongTeks memotong teks agar muat di lebar kolom (dengan "...").
func potongTeks(pdf *gofpdf.Fpdf, teks string, lebar float64) string {
	teks = kartuSanitize(teks)
	if pdf.GetStringWidth(teks) <= lebar-2 {
		return teks
	}
	for len(teks) > 0 && pdf.GetStringWidth(teks+"...") > lebar-2 {
		teks = teks[:len(teks)-1]
	}
	return teks + "..."
}

// hitungHalamanPDF menghitung jumlah halaman PDF dari bytes-nya
// (menghitung "/Type /Page" dikurangi "/Type /Pages"). Dipakai tes.
func hitungHalamanPDF(b []byte) int {
	return bytes.Count(b, []byte("/Type /Page")) - bytes.Count(b, []byte("/Type /Pages"))
}

// ---------------------------------------------------------------------------
// Generator PDF (gofpdf, kompresi DIMATIKAN agar isi teks bisa dicek
// byte stream oleh tes; ukuran PDF kartu tetap kecil).
// ---------------------------------------------------------------------------

func newKartuPDF() *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetCompression(false)
	pdf.SetAutoPageBreak(false, 0)
	pdf.SetMargins(kartuPDFMargin, kartuPDFMargin, kartuPDFMargin)
	// Footer di SETIAP halaman: garis pemisah tipis + catatan + nomor
	// halaman. Posisi konsisten: garis di y=282, teks di y=285.
	pdf.SetFooterFunc(func() {
		pdf.SetDrawColor(200, 200, 200)
		pdf.SetLineWidth(0.3)
		pdf.Line(kartuPDFMargin, 282, kartuPDFMargin+kartuPDFLebarUsable, 282)
		pdf.SetDrawColor(0, 0, 0)
		pdf.SetLineWidth(0.2)
		pdf.SetY(285)
		pdf.SetFont("helvetica", "I", 8)
		pdf.SetTextColor(80, 80, 80)
		pdf.CellFormat(kartuPDFLebarUsable-15, 4, kartuSanitize(kartuPDFCatatan), "", 0, "C", false, 0, "")
		pdf.CellFormat(15, 4, fmt.Sprintf("Hal. %d", pdf.PageNo()), "", 0, "R", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
	})
	return pdf
}

func tulisKopKartu(pdf *gofpdf.Fpdf, data kartuPDFData) {
	pdf.SetY(kartuPDFMargin)
	if strings.TrimSpace(data.NamaSekolah) != "" {
		// Nama sekolah paling menonjol: 16pt tebal, huruf besar.
		pdf.SetFont("helvetica", "B", kartuPDFUkuranNamaSekolah)
		pdf.CellFormat(kartuPDFLebarUsable, 9, kartuSanitize(strings.ToUpper(data.NamaSekolah)), "", 1, "C", false, 0, "")
	}
	// Judul tebal 15pt; tinggi baris tetap menjaga spasi antar blok konsisten.
	pdf.SetFont("helvetica", "B", kartuPDFUkuranJudul)
	pdf.CellFormat(kartuPDFLebarUsable, 9, kartuPDFJudul, "", 1, "C", false, 0, "")
	namaEvent := strings.TrimSpace(data.NamaEvent)
	if namaEvent == "" {
		namaEvent = kartuPDFEventDefault
	}
	pdf.SetFont("helvetica", "", kartuPDFUkuranEvent)
	pdf.SetTextColor(70, 70, 70)
	pdf.CellFormat(kartuPDFLebarUsable, 6.5, kartuSanitize(namaEvent), "", 1, "C", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	// Garis pemisah kop yang tegas: garis biru tebal + garis tipis di bawahnya.
	pdf.SetDrawColor(30, 78, 158)
	pdf.SetLineWidth(0.9)
	y := pdf.GetY() + 1
	pdf.Line(kartuPDFMargin, y, kartuPDFMargin+kartuPDFLebarUsable, y)
	pdf.SetLineWidth(0.3)
	pdf.Line(kartuPDFMargin, y+1.8, kartuPDFMargin+kartuPDFLebarUsable, y+1.8)
	pdf.SetLineWidth(0.2)
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetY(y + 1.8 + 2.5)
}

func tulisIdentitasKartu(pdf *gofpdf.Fpdf, data kartuPDFData) {
	baris := []struct {
		label string
		nilai string
		tebal bool
	}{
		{"Nama", data.NamaSiswa, true},
		{"NISN", data.NISN, false},
		{"Kelas", data.Kelas, false},
		{"Jurusan", data.Jurusan, false},
		{"Username", data.Username, false},
	}
	// Nomor peserta TIDAK ada di blok identitas — nomor tampil per baris
	// tabel jadwal (satu siswa bisa bernomor beda di tiap jadwal).
	// Label kolom lebar tetap + ":" sejajar; tiap baris tinggi sama
	// (7 mm); inset 2 mm dari margin agar teks tidak mepet tepi.
	lebarNilai := kartuPDFLebarUsable - kartuPDFInsetIsi - kartuPDFLebarLabelIdentitas - kartuPDFLebarTitikDua
	for _, b := range baris {
		if b.tebal {
			pdf.SetFont("helvetica", "B", 10)
		} else {
			pdf.SetFont("helvetica", "", 10)
		}
		pdf.SetX(kartuPDFMargin + kartuPDFInsetIsi)
		pdf.CellFormat(kartuPDFLebarLabelIdentitas, kartuPDFTinggiBarisIdent, kartuSanitize(b.label), "", 0, "L", false, 0, "")
		pdf.CellFormat(kartuPDFLebarTitikDua, kartuPDFTinggiBarisIdent, ":", "", 0, "L", false, 0, "")
		pdf.CellFormat(lebarNilai, kartuPDFTinggiBarisIdent, kartuSanitize(b.nilai), "", 1, "L", false, 0, "")
	}
	pdf.Ln(2.5)
}

func tulisHeaderTabelKartu(pdf *gofpdf.Fpdf) {
	// Header tebal, latar abu terang, teks biru tua, garis bawah tegas.
	// Border hanya bawah ("B"): pemisah antar baris tabel berupa garis
	// horizontal, konsisten dengan baris data.
	pdf.SetFont("helvetica", "B", 9)
	pdf.SetFillColor(233, 236, 240)
	pdf.SetTextColor(30, 78, 158)
	pdf.SetDrawColor(30, 78, 158)
	pdf.SetLineWidth(0.6)
	for i, h := range kartuPDFHeaderKolom {
		pdf.CellFormat(kartuPDFLebarKolom[i], kartuPDFTinggiHeaderTabel, h, "B", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.2)
	pdf.SetFillColor(255, 255, 255)
}

func tulisTabelJadwalKartu(pdf *gofpdf.Fpdf, data kartuPDFData) {
	pdf.SetFont("helvetica", "B", 11)
	pdf.CellFormat(kartuPDFLebarUsable, 7, "Jadwal Ujian", "", 1, "L", false, 0, "")
	pdf.Ln(1)
	tulisHeaderTabelKartu(pdf)
	if len(data.Jadwal) == 0 {
		pdf.SetFont("helvetica", "I", 9)
		pdf.CellFormat(kartuPDFLebarUsable, 7, "Belum ada jadwal ujian untuk kelas ini.", "1", 1, "C", false, 0, "")
		return
	}
	pdf.SetFont("helvetica", "", 9)
	// Garis horizontal tiap baris: warna abu terang, lebar 0.2.
	pdf.SetDrawColor(200, 205, 212)
	for idx, j := range data.Jadwal {
		// Pindah halaman bila baris tidak muat; header tabel diulang,
		// identitas TIDAK diulang.
		if pdf.GetY()+kartuPDFTinggiBarisTabel > kartuPDFBatasBawah {
			pdf.AddPage()
			tulisHeaderTabelKartu(pdf)
			pdf.SetFont("helvetica", "", 9)
			pdf.SetDrawColor(200, 205, 212)
		}
		// Zebra striping halus: baris ganjil latar abu sangat terang.
		if idx%2 == 1 {
			pdf.SetFillColor(243, 245, 249)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		sel := []string{
			fmt.Sprintf("%d", idx+1),
			j.NomorPeserta,
			j.Mapel,
			j.Paket,
			j.Tanggal,
			j.Jam,
			j.Ruangan,
		}
		align := []string{"C", "C", "L", "L", "C", "C", "L"}
		for i, v := range sel {
			// Kolom Nomor Peserta (44 mm) TIDAK boleh terpotong: lebar
			// kolom dipertahankan, potongTeks hanya memotong bila
			// benar-benar melebihi kolom.
			pdf.CellFormat(kartuPDFLebarKolom[i], kartuPDFTinggiBarisTabel, potongTeks(pdf, v, kartuPDFLebarKolom[i]), "B", 0, align[i], true, 0, "")
		}
		pdf.Ln(-1)
	}
	pdf.SetFillColor(255, 255, 255)
	pdf.SetDrawColor(0, 0, 0)
	pdf.Ln(3)
}

func tulisKotakLoginKartu(pdf *gofpdf.Fpdf, data kartuPDFData) {
	tinggi := kartuPDFTinggiKotakLogin
	if pdf.GetY()+tinggi > kartuPDFBatasBawah {
		pdf.AddPage()
	}
	x := kartuPDFMargin
	y := pdf.GetY()
	pdf.SetDrawColor(30, 78, 158)
	pdf.SetDashPattern([]float64{2, 2}, 0)
	pdf.Rect(x, y, kartuPDFLebarUsable, tinggi, "D")
	pdf.SetDashPattern([]float64{}, 0)
	pdf.SetDrawColor(0, 0, 0)
	// Judul kotak.
	pdf.SetXY(x+kartuPDFInsetIsi+2, y+2.5)
	pdf.SetFont("helvetica", "B", 10)
	pdf.CellFormat(kartuPDFLebarUsable-2*(kartuPDFInsetIsi+2), kartuPDFTinggiBarisLogin, "Login Ujian", "", 1, "L", false, 0, "")
	// Dua baris label + nilai sejajar (label lebar tetap 40 mm + ":").
	// Sebelumnya satu baris gabungan yang mepet lebar kotak; kini tiap
	// kredensial punya barisnya sendiri sehingga tidak terpotong.
	baris := []struct {
		label string
		nilai string
	}{
		{"Username", data.Username},
		{"Kata sandi awal", kartuSanitize(data.NISN) + " (wajib diganti saat login pertama)"},
	}
	lebarNilai := kartuPDFLebarUsable - 2*(kartuPDFInsetIsi+2) - kartuPDFLebarLabelLogin - kartuPDFLebarTitikDua
	for _, b := range baris {
		pdf.SetX(x + kartuPDFInsetIsi + 2)
		pdf.SetFont("helvetica", "", 10)
		pdf.CellFormat(kartuPDFLebarLabelLogin, kartuPDFTinggiBarisLogin, kartuSanitize(b.label), "", 0, "L", false, 0, "")
		pdf.CellFormat(kartuPDFLebarTitikDua, kartuPDFTinggiBarisLogin, ":", "", 0, "L", false, 0, "")
		pdf.SetFont("helvetica", "B", 10)
		pdf.CellFormat(lebarNilai, kartuPDFTinggiBarisLogin, kartuSanitize(b.nilai), "", 1, "L", false, 0, "")
	}
	pdf.SetY(y + tinggi + 3)
}

func tulisTandaTanganKartu(pdf *gofpdf.Fpdf, data kartuPDFData) {
	tinggi := 32.0
	if pdf.GetY()+tinggi > kartuPDFBatasBawah {
		pdf.AddPage()
	}
	y0 := pdf.GetY()
	lebarKolom := kartuPDFLebarUsable / 2
	// Dua kolom sejajar: Peserta Ujian (kiri) | Panitia (kanan).
	pdf.SetFont("helvetica", "", 10)
	pdf.SetXY(kartuPDFMargin, y0)
	pdf.CellFormat(lebarKolom, 6, "Peserta Ujian", "", 0, "C", false, 0, "")
	pdf.CellFormat(lebarKolom, 6, "Panitia", "", 1, "C", false, 0, "")
	// Ruang tanda tangan + garis: kedua garis SAMA PANJANG (66 mm),
	// inset 12 mm dari tiap sisi kolomnya.
	yGaris := y0 + 6 + kartuPDFRuangTTD
	pdf.SetLineWidth(0.4)
	pdf.Line(kartuPDFMargin+kartuPDFInsetGarisTTD, yGaris, kartuPDFMargin+lebarKolom-kartuPDFInsetGarisTTD, yGaris)
	pdf.Line(kartuPDFMargin+lebarKolom+kartuPDFInsetGarisTTD, yGaris, kartuPDFMargin+kartuPDFLebarUsable-kartuPDFInsetGarisTTD, yGaris)
	pdf.SetLineWidth(0.2)
	// Nama terang tepat 2 mm di bawah garis, di kedua kolom.
	pdf.SetXY(kartuPDFMargin, yGaris+2)
	pdf.CellFormat(lebarKolom, 6, kartuSanitize("( "+data.NamaSiswa+" )"), "", 0, "C", false, 0, "")
	pdf.CellFormat(lebarKolom, 6, "( Nama Terang )", "", 1, "C", false, 0, "")
	pdf.SetY(yGaris + 2 + 6 + 2)
}

// tulisSatuKartu menulis satu kartu lengkap di halaman baru.
// Dipanggil setelah AddPage oleh pemanggil (atau halaman pertama).
func tulisSatuKartu(pdf *gofpdf.Fpdf, data kartuPDFData) {
	tulisKopKartu(pdf, data)
	tulisIdentitasKartu(pdf, data)
	tulisTabelJadwalKartu(pdf, data)
	tulisKotakLoginKartu(pdf, data)
	tulisTandaTanganKartu(pdf, data)
}

// renderKartuPDFBytes menutup PDF dan mengembalikan bytes-nya.
func renderKartuPDFBytes(pdf *gofpdf.Fpdf) ([]byte, error) {
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderKartuPDF membuat PDF 1 kartu untuk satu siswa.
func RenderKartuPDF(data kartuPDFData) ([]byte, error) {
	pdf := newKartuPDF()
	pdf.AddPage()
	tulisSatuKartu(pdf, data)
	return renderKartuPDFBytes(pdf)
}

// RenderKartuBulkPDF membuat PDF gabungan: 1 halaman (awal) per siswa.
func RenderKartuBulkPDF(cards []kartuPDFData) ([]byte, error) {
	if len(cards) == 0 {
		return nil, fmt.Errorf("tidak ada kartu untuk dibuat PDF")
	}
	if len(cards) > kartuPDFMaksSiswaBulk {
		return nil, errKartuTerlaluBanyak
	}
	pdf := newKartuPDF()
	for _, c := range cards {
		pdf.AddPage()
		tulisSatuKartu(pdf, c)
	}
	return renderKartuPDFBytes(pdf)
}

// ---------------------------------------------------------------------------
// Loader data kartu PDF (DB).
// ---------------------------------------------------------------------------

// loadJadwalKelasKartu memuat SEMUA jadwal kelas siswa, urut tanggal + jam.
// Filter event opsional (eventID == "" = semua event). Status jadwal TIDAK
// difilter: kartu memuat semua jadwal kelas sesuai spesifikasi.
func loadJadwalKelasKartu(ctx context.Context, pool *pgxpool.Pool, kelasID, eventID string) ([]jadwalKartu, error) {
	q := `SELECT j.id, j.kelas_id, to_char(j.tanggal, 'YYYY-MM-DD'),
		        to_char(j.jam_mulai, 'HH24:MI'), to_char(j.jam_selesai, 'HH24:MI'), j.status,
		        p.nama, p.jenis, m.nama, k.nama, ju.nama,
		        j.event_id, e.kode, e.nama, j.ruangan_id, r.nama, r.lokasi
		 FROM jadwal_ujian j
		 JOIN paket_ujian p ON p.id = j.paket_id
		 JOIN mapel m ON m.id = p.mapel_id
		 JOIN kelas k ON k.id = j.kelas_id
		 JOIN jurusan ju ON ju.id = k.jurusan_id
		 LEFT JOIN event e ON e.id = j.event_id
		 LEFT JOIN ruangan r ON r.id = j.ruangan_id
		 WHERE j.kelas_id = $1`
	args := []any{kelasID}
	if eventID != "" {
		q += ` AND j.event_id = $2`
		args = append(args, eventID)
	}
	q += ` ORDER BY j.tanggal ASC, j.jam_mulai ASC`
	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []jadwalKartu{}
	for rows.Next() {
		var j jadwalKartu
		if err := rows.Scan(&j.ID, &j.KelasID, &j.Tanggal, &j.JamMulai, &j.JamSelesai, &j.Status,
			&j.PaketNama, &j.PaketJenis, &j.MapelNama, &j.KelasNama, &j.JurusanNama,
			&j.EventID, &j.EventKode, &j.EventNama, &j.RuanganID, &j.RuanganNama, &j.RuanganLokasi); err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

// loadKartuPDFData menyusun satu kartu PDF untuk siswa (by siswa.id).
// found=false bila siswa tidak ada / tidak aktif (mirip endpoint JSON:
// siswa tidak aktif tidak punya kartu). eventIDFilter membatasi jadwal
// yang tampil ("" = semua jadwal kelas).
func loadKartuPDFData(ctx context.Context, pool *pgxpool.Pool, siswaID, eventIDFilter string) (*kartuPDFData, bool, error) {
	var d kartuPDFData
	var kelasID string
	var isActive bool
	var userID *string
	err := pool.QueryRow(ctx,
		`SELECT s.nisn, s.nama, s.kelas_id, s.user_id, s.is_active,
		        k.nama, ju.nama, COALESCE(u.username, s.nisn)
		 FROM siswa s
		 JOIN kelas k ON k.id = s.kelas_id
		 JOIN jurusan ju ON ju.id = k.jurusan_id
		 LEFT JOIN users u ON u.id = s.user_id
		 WHERE s.id = $1`, siswaID).
		Scan(&d.NISN, &d.NamaSiswa, &kelasID, &userID, &isActive, &d.Kelas, &d.Jurusan, &d.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if !isActive {
		return nil, false, nil
	}
	jadwalList, err := loadJadwalKelasKartu(ctx, pool, kelasID, eventIDFilter)
	if err != nil {
		return nil, false, err
	}
	for _, j := range jadwalList {
		// Nomor peserta per jadwal dari buildKartuList (logika penomoran
		// TIDAK diduplikasi di sini).
		items, err := buildKartuList(ctx, pool, &j)
		if err != nil {
			return nil, false, err
		}
		nomor := ""
		for _, k := range items {
			if k.SiswaID == siswaID {
				nomor = k.NomorPeserta
				break
			}
		}
		// Ruangan tampil nama saja (tanpa lokasi) agar muat di kolom
		// sempit; lokasi ada di data master ruangan bila dibutuhkan.
		ruangan := "-"
		if j.RuanganNama != nil && strings.TrimSpace(*j.RuanganNama) != "" {
			ruangan = strings.TrimSpace(*j.RuanganNama)
		}
		eventNama := ""
		if j.EventNama != nil {
			eventNama = *j.EventNama
		}
		d.Jadwal = append(d.Jadwal, kartuPDFJadwal{
			NomorPeserta: nomor,
			Mapel:        j.MapelNama,
			Paket:        j.PaketNama,
			Tanggal:      formatTanggalKartu(j.Tanggal),
			Jam:          j.JamMulai + " - " + j.JamSelesai,
			Ruangan:      ruangan,
			EventNama:    eventNama,
		})
	}
	d.NamaSekolah = loadNamaSekolah(ctx, pool)
	d.NamaEvent = namaEventGabungan(d.Jadwal)
	return &d, true, nil
}

// loadKartuBulkPDFData menyusun kartu semua siswa aktif (urut nama A-Z),
// opsional filter kelas_id / event_id. Melebihi batas -> errKartuTerlaluBanyak.
func loadKartuBulkPDFData(ctx context.Context, pool *pgxpool.Pool, kelasID, eventID string) ([]kartuPDFData, error) {
	q := `SELECT s.id FROM siswa s WHERE s.is_active = TRUE`
	args := []any{}
	if kelasID != "" {
		args = append(args, kelasID)
		q += fmt.Sprintf(` AND s.kelas_id = $%d`, len(args))
	}
	// Filter event: hanya siswa yang kelasnya punya jadwal di event itu.
	if eventID != "" {
		args = append(args, eventID)
		q += fmt.Sprintf(` AND EXISTS (SELECT 1 FROM jadwal_ujian j
			WHERE j.kelas_id = s.kelas_id AND j.event_id = $%d)`, len(args))
	}
	q += ` ORDER BY s.nama ASC`
	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) > kartuPDFMaksSiswaBulk {
		return nil, errKartuTerlaluBanyak
	}
	cards := []kartuPDFData{}
	for _, id := range ids {
		d, found, err := loadKartuPDFData(ctx, pool, id, eventID)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		cards = append(cards, *d)
	}
	// Urutan nama A-Z sudah dari query; jaga deterministik bila ada
	// duplikat nama.
	sort.SliceStable(cards, func(i, j int) bool { return cards[i].NamaSiswa < cards[j].NamaSiswa })
	return cards, nil
}

// ---------------------------------------------------------------------------
// Handler HTTP.
// ---------------------------------------------------------------------------

// handleKartuSiswaPDF: GET /api/v1/kartu/siswa/:siswaId/pdf
// Siswa hanya kartunya sendiri (bukan pemilik -> 404, anti-IDOR);
// staf admin/proktor/guru boleh semua (gate requireKartuStaf).
func handleKartuSiswaPDF(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c *echo.Context) error {
		uid, role, _, _, ok := audit.CurrentUser(c)
		if !ok {
			return platform.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
		}
		ctx := c.Request().Context()
		s, found, err := loadSiswaKartu(ctx, pool, c.Param("siswaId"))
		if err != nil {
			return platform.WriteError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca kartu peserta.")
		}
		if !found {
			return platform.WriteError(c, http.StatusNotFound, "not_found", "Kartu peserta tidak ditemukan.")
		}
		if role == "siswa" {
			milikSendiri := (s.UserID != nil && *s.UserID == uid) || s.ID == uid
			if !milikSendiri {
				return platform.WriteError(c, http.StatusNotFound, "not_found", "Kartu peserta tidak ditemukan.")
			}
		} else if _, _, ok := requireKartuStaf(c); !ok {
			return nil
		}
		data, found, err := loadKartuPDFData(ctx, pool, s.ID, "")
		if err != nil {
			return platform.WriteError(c, http.StatusInternalServerError, "internal_error", "Gagal menyusun kartu peserta.")
		}
		if !found {
			return platform.WriteError(c, http.StatusNotFound, "not_found", "Kartu peserta tidak ditemukan.")
		}
		pdfBytes, err := RenderKartuPDF(*data)
		if err != nil {
			return platform.WriteError(c, http.StatusInternalServerError, "internal_error", "Gagal membuat PDF kartu peserta.")
		}
		c.Response().Header().Set("Content-Type", "application/pdf")
		c.Response().Header().Set("Content-Disposition",
			`attachment; filename="`+namaFileKartuPDF(data.NISN)+`"`)
		return c.Blob(http.StatusOK, "application/pdf", pdfBytes)
	}
}

// handleKartuBulkPDF: GET /api/v1/kartu/pdf?kelas_id=&event_id=
// Staf admin/proktor/guru saja (gate SAMA seperti endpoint kartu lama).
func handleKartuBulkPDF(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c *echo.Context) error {
		if _, _, ok := requireKartuStaf(c); !ok {
			return nil
		}
		ctx := c.Request().Context()
		kelasID := strings.TrimSpace(c.QueryParam("kelas_id"))
		eventID := strings.TrimSpace(c.QueryParam("event_id"))
		if kelasID != "" && !isValidUUID(kelasID) {
			return platform.WriteError(c, http.StatusBadRequest, "validation_error", "kelas_id harus UUID yang valid.")
		}
		if eventID != "" && !isValidUUID(eventID) {
			return platform.WriteError(c, http.StatusBadRequest, "validation_error", "event_id harus UUID yang valid.")
		}
		if kelasID != "" {
			var ada int
			if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM kelas WHERE id = $1`, kelasID).Scan(&ada); err != nil {
				return platform.WriteError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa kelas.")
			}
			if ada == 0 {
				return platform.WriteError(c, http.StatusNotFound, "not_found", "Kelas tidak ditemukan.")
			}
		}
		if eventID != "" {
			if _, found, err := loadEvent(ctx, pool, eventID); err != nil {
				return platform.WriteError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa event.")
			} else if !found {
				return platform.WriteError(c, http.StatusNotFound, "not_found", "Event tidak ditemukan.")
			}
		}
		cards, err := loadKartuBulkPDFData(ctx, pool, kelasID, eventID)
		if err != nil {
			if errors.Is(err, errKartuTerlaluBanyak) {
				return platform.WriteError(c, http.StatusBadRequest, "too_many_siswa",
					fmt.Sprintf("Jumlah siswa melebihi batas %d — gunakan filter kelas_id / event_id.", kartuPDFMaksSiswaBulk))
			}
			return platform.WriteError(c, http.StatusInternalServerError, "internal_error", "Gagal menyusun kartu peserta.")
		}
		if len(cards) == 0 {
			return platform.WriteError(c, http.StatusNotFound, "not_found", "Tidak ada kartu peserta untuk filter tersebut.")
		}
		pdfBytes, err := RenderKartuBulkPDF(cards)
		if err != nil {
			if errors.Is(err, errKartuTerlaluBanyak) {
				return platform.WriteError(c, http.StatusBadRequest, "too_many_siswa",
					fmt.Sprintf("Jumlah siswa melebihi batas %d — gunakan filter kelas_id / event_id.", kartuPDFMaksSiswaBulk))
			}
			return platform.WriteError(c, http.StatusInternalServerError, "internal_error", "Gagal membuat PDF kartu peserta.")
		}
		c.Response().Header().Set("Content-Type", "application/pdf")
		c.Response().Header().Set("Content-Disposition", `attachment; filename="kartu-peserta-gabungan.pdf"`)
		return c.Blob(http.StatusOK, "application/pdf", pdfBytes)
	}
}
