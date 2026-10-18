// Package analisis — Analisis Butir Soal + Nilai per Kompetensi (Tim Fase 2).
//
// Dihitung ON-DEMAND dari tabel yang sudah ada (jadwal_ujian, attempt,
// jawaban, soal, kompetensi, siswa) — TANPA tabel cache.
//
// Definisi FINAL (spesifikasi tugas Fase 2, mengikat; beda ambang dengan
// ANALISIS.md Bab 8.4 yang lebih longgar — yang dipakai adalah definisi
// di bawah ini):
//
//   - Populasi: attempt yang SUDAH selesai pada satu jadwal
//     (attempt.status = 'selesai', atau punya baris nilai_akhir final).
//     Attempt 'berjalan' / 'diblokir' TIDAK dihitung.
//   - Tingkat kesukaran P = jumlah benar / jumlah penjawab.
//     Label: P < 0.3 Sukar; 0.3–0.7 Sedang; P > 0.7 Mudah.
//   - Daya beda D = (benar kelompok atas − benar kelompok bawah) / n,
//     n = ukuran tiap kelompok (kelompok sama besar). Kelompok atas/bawah
//     = peserta ber-nilai tertinggi/terendah pada jadwal itu.
//     N >= 10: n = round(27% × N) (metode 27%).
//     N < 10: n = floor(N/2) (median split 50%) + catatan "sampel kecil".
//     Interpretasi: D >= 0.40 Sangat Baik; 0.30–0.39 Baik; 0.20–0.29
//     Cukup; 0.00–0.19 Buruk; D < 0.00 Sangat Buruk (soal bermasalah).
//   - Pengecoh: per opsi jawaban, jumlah pemilih + persentase dari
//     penjawab soal itu. Pengecoh "tidak berfungsi" bila dipilih < 5%.
//     Hanya untuk soal berbasis opsi (PG / PG Kompleks / dsb.); untuk
//     soal lain pengecoh = N/A.
//   - Nilai per kompetensi (per siswa, per jadwal):
//     nilai = Σ poin / Σ bobot × 100 atas soal kompetensi itu, dengan:
//   - soal tidak dijawab = poin 0, bobot TETAP masuk penyebut
//     (sama seperti nilai akhir di grading.HitungNilaiAkhir);
//   - soal manual yang SUDAH dijawab tapi BELUM dikoreksi = "belum
//     dinilai" (ANALISIS Bab 8.2): poin tidak masuk pembilang,
//     bobot tidak masuk penyebut;
//   - siswa tidak menjawab satu pun soal kompetensi itu ->
//     nilai = null + flag "tidak ada soal" (BUKAN 0).
//
// "Benar" untuk satu jawaban:
//   - Soal otomatis: poin == bobot (poin penuh). Berlaku juga untuk
//     soal proporsional (PG Kompleks, Menjodohkan, dsb.): hanya poin
//     penuh yang dihitung benar.
//   - Soal manual (uraian / upload_file / praktik_singkat): poin >= 0.6 ×
//     bobot (ambang lulus per soal).
//
// Urutan nilai untuk pemeringkatan kelompok: nilai_akhir.nilai bila ada,
// bila tidak ada dihitung dari jawaban.poin (Σ poin / Σ bobot paket).
//
// Endpoint (Register):
//
//	GET /api/v1/analisis/:jadwalId            (permission analisis.read)
//	GET /api/v1/analisis/:jadwalId/export.csv (permission analisis.export)
//	GET /api/v1/kompetensi/:jadwalId          (permission rekap.read)
//	GET /api/v1/kompetensi/:jadwalId/export.csv (permission rekap.export)
//
// Nilai per kompetensi memakai permission rekap.* (sesuai spesifikasi
// tugas) karena ia adalah laporan rekap, bukan analisis butir.
package analisis

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/grading"

	"github.com/google/uuid"
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

// requirePerm: 401 bila belum login (context kosong), 403 bila tidak
// punya permission — sesuai KONTRAK.md §1.2.
func requirePerm(c *echo.Context, perm string) (userID, role, siswaID string, ok bool) {
	var okUser bool
	userID, role, _, siswaID, okUser = audit.CurrentUser(c)
	if !okUser {
		return "", "", "", false
	}
	if !audit.HasPermission(c, perm) {
		return "", "", "", false
	}
	return userID, role, siswaID, true
}

func round2(x float64) float64 { return math.Round(x*100) / 100 }

// SembunyikanPertanyaanUntukProktor mengosongkan field Pertanyaan pada
// tiap item bila role pemanggil = proktor (Fase 3, SoD proktor vs isi
// soal — KONTRAK §5.3: proktor TIDAK BOLEH melihat isi soal, hanya
// metadata/statistik). Statistik (kesukaran, daya beda, pengecoh) tetap
// tampil. Guru/admin tetap melihat pertanyaan. Fungsi murni agar bisa
// diuji tanpa DB.
func SembunyikanPertanyaanUntukProktor(hasil analisisResult, role string) analisisResult {
	if role != "proktor" {
		return hasil
	}
	// Salin slice agar data pemanggil tidak termutasi.
	salinan := hasil
	salinan.Items = append([]analisisItem(nil), hasil.Items...)
	for i := range salinan.Items {
		salinan.Items[i].Pertanyaan = ""
	}
	return salinan
}

// ---------------------------------------------------------------------------
// Input mentah (dipakai fungsi murni di bawah, agar bisa diuji tanpa DB).
// ---------------------------------------------------------------------------

// opsiSoal adalah satu opsi jawaban dari soal.jawaban ({"opsi": [...]}).
type opsiSoal struct {
	Label string
	Teks  string
}

// soalInput adalah satu soal paket ujian.
type soalInput struct {
	ID             string
	Urutan         int
	Pertanyaan     string
	Tipe           string
	Bobot          float64
	KompetensiID   *string
	KompetensiKode string
	KompetensiNama string
	Opsi           []opsiSoal
}

// jawabanInput adalah satu baris tabel jawaban.
type jawabanInput struct {
	AttemptID string
	SoalID    string
	// Jawaban nil = soal tidak dijawab (baris tidak ada / jawaban NULL).
	Jawaban   json.RawMessage
	Poin      *float64 // nil = belum dikoreksi (soal manual) / belum dinilai
	Dikoreksi bool
}

// attemptInput adalah satu attempt selesai.
type attemptInput struct {
	ID      string
	SiswaID string
	// Nilai dari nilai_akhir bila ada; nil = hitung dari jawaban.poin.
	Nilai *float64
}

// siswaInput adalah identitas siswa untuk satu attempt.
type siswaInput struct {
	AttemptID string
	SiswaID   string
	NISN      string
	Nama      string
	Kelas     string
}

// ---------------------------------------------------------------------------
// Fungsi murni (tanpa DB) — diuji di analisis_test.go.
// ---------------------------------------------------------------------------

// ringkas memadatkan whitespace dan memotong pertanyaan ke maxRunes rune.
func ringkas(s string, maxRunes int) string {
	s = strings.Join(strings.Fields(s), " ")
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	r := []rune(s)
	return strings.TrimSpace(string(r[:maxRunes])) + "…"
}

// isBenar menilai satu jawaban menurut definisi "benar" di atas.
// Mengembalikan false bila jawaban tidak dijawab / belum dikoreksi.
func isBenar(tipe string, j jawabanInput, bobot float64) bool {
	if j.Jawaban == nil || len(strings.TrimSpace(string(j.Jawaban))) == 0 ||
		strings.TrimSpace(string(j.Jawaban)) == "null" {
		return false
	}
	if j.Poin == nil {
		return false
	}
	if grading.Manual(tipe) {
		return *j.Poin >= 0.6*bobot
	}
	return math.Abs(*j.Poin-bobot) < 1e-9
}

// dijawab melaporkan apakah soal dijawab (jawaban non-NULL).
func dijawab(j jawabanInput) bool {
	return j.Jawaban != nil && len(strings.TrimSpace(string(j.Jawaban))) > 0 &&
		strings.TrimSpace(string(j.Jawaban)) != "null"
}

// pilihanTerpilih mengekstrak label opsi yang dipilih siswa dari jawaban.
// Bentuk yang diterima (toleran, mengikuti format frontend + grading):
// "B", ["A","C"], {"pilihan":"B"}, {"jawaban":"B"}, {"opsi":"B"},
// {"opsi":["A","C"]}, {"jawaban":["A","C"]}.
func pilihanTerpilih(raw json.RawMessage) []string {
	if len(strings.TrimSpace(string(raw))) == 0 || strings.TrimSpace(string(raw)) == "null" {
		return nil
	}
	if s, ok := parseString(raw); ok {
		if s == "" {
			return nil
		}
		return []string{s}
	}
	if arr, ok := parseStringSlice(raw); ok {
		return arr
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil
	}
	for _, key := range []string{"opsi", "pilihan", "jawaban"} {
		if v, ada := obj[key]; ada {
			if s, ok := parseString(v); ok && s != "" {
				return []string{s}
			}
			if arr, ok := parseStringSlice(v); ok {
				return arr
			}
		}
	}
	return nil
}

func parseString(raw []byte) (string, bool) {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, true
	}
	return "", false
}

func parseStringSlice(raw []byte) ([]string, bool) {
	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, true
	}
	if s, ok := parseString(raw); ok && s != "" {
		return []string{s}, true
	}
	return nil, false
}

// labelKesukaran: <0.3 Sukar, 0.3–0.7 Sedang, >0.7 Mudah.
func labelKesukaran(p float64) string {
	switch {
	case p < 0.3:
		return "Sukar"
	case p <= 0.7:
		return "Sedang"
	default:
		return "Mudah"
	}
}

// interpretasiDayaBeda dari nilai MENTAH (belum dibulatkan) agar
// pembulatan tampilan tidak menggeser batas interpretasi.
func interpretasiDayaBeda(d float64) string {
	switch {
	case d >= 0.40:
		return "Sangat Baik"
	case d >= 0.30:
		return "Baik"
	case d >= 0.20:
		return "Cukup"
	case d >= 0.00:
		return "Buruk"
	default:
		return "Sangat Buruk"
	}
}

// opsiBased: soal berbasis opsi (pengecoh bermakna) atau bukan.
func opsiBased(tipe string) bool {
	switch grading.NormalisasiTipe(tipe) {
	case "pg", "pg_kompleks":
		return true
	default:
		return false
	}
}

// nilaiAttempt menghitung nilai 0–100 satu attempt untuk pemeringkatan:
// nilai_akhir bila ada, else Σ poin / Σ bobot paket (jawaban tanpa poin = 0).
func nilaiAttempt(a attemptInput, jawabanPerAttempt map[string][]jawabanInput, totalBobot float64) float64 {
	if a.Nilai != nil {
		return *a.Nilai
	}
	if totalBobot <= 0 {
		return 0
	}
	sum := 0.0
	for _, j := range jawabanPerAttempt[a.ID] {
		if j.Poin != nil {
			sum += *j.Poin
		}
	}
	return sum / totalBobot * 100
}

// --- Output analisis butir ---

type pengecohOut struct {
	Opsi      string  `json:"opsi"`
	Teks      string  `json:"teks,omitempty"`
	Pemilih   int     `json:"pemilih"`
	Persen    float64 `json:"persen"`
	Berfungsi bool    `json:"berfungsi"`
}

type analisisItem struct {
	SoalID               string        `json:"soal_id"`
	Urutan               int           `json:"urutan"`
	Pertanyaan           string        `json:"pertanyaan"`
	Tipe                 string        `json:"tipe"`
	Bobot                float64       `json:"bobot"`
	KompetensiID         *string       `json:"kompetensi_id"`
	KompetensiKode       string        `json:"kompetensi_kode"`
	Kompetensi           string        `json:"kompetensi"`
	JumlahPenjab         int           `json:"jumlah_penjab"`
	JumlahBenar          int           `json:"jumlah_benar"`
	JumlahTidakDijawab   int           `json:"jumlah_tidak_dijawab"`
	BelumDikoreksi       int           `json:"belum_dikoreksi"`
	Kesukaran            *float64      `json:"kesukaran"`
	LabelKesukaran       string        `json:"label_kesukaran"`
	DayaBeda             *float64      `json:"daya_beda"`
	InterpretasiDayaBeda string        `json:"interpretasi_daya_beda"`
	Pengecoh             []pengecohOut `json:"pengecoh"`
	PengecohStatus       string        `json:"pengecoh_status"` // "ada" | "n/a"
}

type analisisResult struct {
	JadwalID       string         `json:"jadwal_id"`
	JumlahPeserta  int            `json:"jumlah_peserta"`
	JumlahSoal     int            `json:"jumlah_soal"`
	MetodeKelompok string         `json:"metode_kelompok"`
	SampelKecil    bool           `json:"sampel_kecil"`
	Catatan        string         `json:"catatan"`
	Items          []analisisItem `json:"items"`
}

// computeAnalisis menghitung analisis butir dari data mentah.
func computeAnalisis(jadwalID string, soals []soalInput, attempts []attemptInput, jawabans []jawabanInput) analisisResult {
	soalByID := map[string]soalInput{}
	totalBobot := 0.0
	for _, s := range soals {
		soalByID[s.ID] = s
		totalBobot += s.Bobot
	}
	jawabanPerAttempt := map[string][]jawabanInput{}
	jawabanPerSoal := map[string]map[string]jawabanInput{}
	for _, j := range jawabans {
		if _, ok := soalByID[j.SoalID]; !ok {
			continue // jawaban untuk soal di luar paket: abaikan
		}
		jawabanPerAttempt[j.AttemptID] = append(jawabanPerAttempt[j.AttemptID], j)
		if jawabanPerSoal[j.SoalID] == nil {
			jawabanPerSoal[j.SoalID] = map[string]jawabanInput{}
		}
		jawabanPerSoal[j.SoalID][j.AttemptID] = j
	}

	// Pemeringkatan kelompok atas/bawah berdasarkan nilai akhir.
	type ranked struct {
		id    string
		nilai float64
	}
	rankedList := make([]ranked, 0, len(attempts))
	for _, a := range attempts {
		rankedList = append(rankedList, ranked{a.ID, nilaiAttempt(a, jawabanPerAttempt, totalBobot)})
	}
	sort.SliceStable(rankedList, func(i, j int) bool {
		if rankedList[i].nilai != rankedList[j].nilai {
			return rankedList[i].nilai > rankedList[j].nilai
		}
		return rankedList[i].id < rankedList[j].id
	})

	n := len(attempts)
	metode := ""
	sampelKecil := false
	catatan := ""
	ukuranKelompok := 0
	if n >= 10 {
		ukuranKelompok = int(math.Round(0.27 * float64(n)))
		metode = "27%"
	} else if n >= 2 {
		ukuranKelompok = n / 2
		metode = "median split 50%"
		sampelKecil = true
		catatan = "sampel kecil: peserta < 10, kelompok atas/bawah memakai median split 50%; daya beda tidak dapat dipercaya sepenuhnya"
	} else {
		metode = "tidak dapat dihitung (peserta < 2)"
		sampelKecil = true
		catatan = "sampel kecil: peserta < 2, daya beda tidak dapat dihitung"
	}
	atas := map[string]bool{}
	bawah := map[string]bool{}
	for i := 0; i < ukuranKelompok && i < len(rankedList); i++ {
		atas[rankedList[i].id] = true
	}
	for i := 0; i < ukuranKelompok && i < len(rankedList); i++ {
		bawah[rankedList[len(rankedList)-1-i].id] = true
	}

	items := make([]analisisItem, 0, len(soals))
	for _, s := range soals {
		perAttempt := jawabanPerSoal[s.ID]
		manual := grading.Manual(s.Tipe)
		jumlahBenar, jumlahPenjab, belumDikoreksi, dijawabCount := 0, 0, 0, 0
		benarAtas, benarBawah := 0, 0
		for _, a := range attempts {
			j, ada := perAttempt[a.ID]
			if !ada || !dijawab(j) {
				continue
			}
			dijawabCount++
			if manual && j.Poin == nil {
				belumDikoreksi++
				continue // belum dinilai: tidak masuk penjawab maupun benar
			}
			jumlahPenjab++
			if isBenar(s.Tipe, j, s.Bobot) {
				jumlahBenar++
				if atas[a.ID] {
					benarAtas++
				}
				if bawah[a.ID] {
					benarBawah++
				}
			}
		}

		var kesukaran *float64
		label := "Belum ada jawaban"
		if jumlahPenjab > 0 {
			p := float64(jumlahBenar) / float64(jumlahPenjab)
			r := round2(p)
			kesukaran = &r
			label = labelKesukaran(p)
		}

		var dayaBeda *float64
		interpretasi := "Tidak dapat dihitung"
		if ukuranKelompok > 0 {
			d := float64(benarAtas-benarBawah) / float64(ukuranKelompok)
			r := round2(d)
			dayaBeda = &r
			interpretasi = interpretasiDayaBeda(d)
		}

		pengecoh := []pengecohOut{}
		pengecohStatus := "n/a"
		if opsiBased(s.Tipe) {
			pengecohStatus = "ada"
			for _, o := range s.Opsi {
				pemilih := 0
				for _, a := range attempts {
					j, ada := perAttempt[a.ID]
					if !ada || !dijawab(j) {
						continue
					}
					for _, pilih := range pilihanTerpilih(j.Jawaban) {
						if pilih == o.Label {
							pemilih++
							break
						}
					}
				}
				persen := 0.0
				if dijawabCount > 0 {
					persen = round2(float64(pemilih) / float64(dijawabCount) * 100)
				}
				pengecoh = append(pengecoh, pengecohOut{
					Opsi: o.Label, Teks: o.Teks, Pemilih: pemilih,
					Persen: persen, Berfungsi: persen >= 5,
				})
			}
		}

		items = append(items, analisisItem{
			SoalID: s.ID, Urutan: s.Urutan, Pertanyaan: ringkas(s.Pertanyaan, 120),
			Tipe: s.Tipe, Bobot: s.Bobot,
			KompetensiID: s.KompetensiID, KompetensiKode: s.KompetensiKode,
			Kompetensi:   s.KompetensiNama,
			JumlahPenjab: jumlahPenjab, JumlahBenar: jumlahBenar,
			JumlahTidakDijawab: n - dijawabCount, BelumDikoreksi: belumDikoreksi,
			Kesukaran: kesukaran, LabelKesukaran: label,
			DayaBeda: dayaBeda, InterpretasiDayaBeda: interpretasi,
			Pengecoh: pengecoh, PengecohStatus: pengecohStatus,
		})
	}

	return analisisResult{
		JadwalID: jadwalID, JumlahPeserta: n, JumlahSoal: len(soals),
		MetodeKelompok: metode, SampelKecil: sampelKecil, Catatan: catatan,
		Items: items,
	}
}

// --- Output nilai per kompetensi ---

type kompetensiNilai struct {
	KompetensiID   *string  `json:"kompetensi_id"`
	KompetensiKode string   `json:"kompetensi_kode"`
	Kompetensi     string   `json:"kompetensi"`
	Nilai          *float64 `json:"nilai"`
	JumlahSoal     int      `json:"jumlah_soal"`
	Dijawab        int      `json:"dijawab"`
	BelumDikoreksi int      `json:"belum_dikoreksi"`
	TidakAdaSoal   bool     `json:"tidak_ada_soal"`
	Flag           string   `json:"flag"`
}

type kompetensiSiswa struct {
	SiswaID            string            `json:"siswa_id"`
	NISN               string            `json:"nisn"`
	Nama               string            `json:"nama"`
	Kelas              string            `json:"kelas"`
	NilaiPerKompetensi []kompetensiNilai `json:"nilai_per_kompetensi"`
	RataRata           *float64          `json:"rata_rata"`
}

type kompetensiKelas struct {
	KompetensiID   *string  `json:"kompetensi_id"`
	KompetensiKode string   `json:"kompetensi_kode"`
	Kompetensi     string   `json:"kompetensi"`
	RataRata       *float64 `json:"rata_rata"`
	JumlahSiswa    int      `json:"jumlah_siswa"`
}

type kompetensiResult struct {
	JadwalID              string            `json:"jadwal_id"`
	JumlahSiswa           int               `json:"jumlah_siswa"`
	RataRataKelas         *float64          `json:"rata_rata_kelas"`
	RataRataPerKompetensi []kompetensiKelas `json:"rata_rata_per_kompetensi"`
	Siswa                 []kompetensiSiswa `json:"siswa"`
}

// nilaiKompetensi menghitung nilai satu siswa untuk satu kelompok soal:
// Σ poin (soal ternilai) / (Σ bobot soal ternilai + Σ bobot soal tidak
// dijawab) × 100. Soal manual yang sudah dijawab tapi belum dikoreksi
// dikeluarkan dari pembilang DAN penyebut ("belum dinilai", bukan 0).
func nilaiKompetensi(soalGroup []soalInput, perAttempt map[string]jawabanInput, attemptID string) (nilai *float64, dijawabN, belumDikoreksi int, flag string, tidakAdaSoal bool) {
	var sumPoin, penyebut float64
	scored := 0
	for _, s := range soalGroup {
		j, ada := perAttempt[s.ID]
		if !ada || !dijawab(j) {
			penyebut += s.Bobot // tidak dijawab = poin 0, bobot tetap dihitung
			continue
		}
		dijawabN++
		if grading.Manual(s.Tipe) && j.Poin == nil {
			belumDikoreksi++
			continue // belum dinilai: tidak masuk pembilang maupun penyebut
		}
		scored++
		penyebut += s.Bobot
		if j.Poin != nil {
			sumPoin += *j.Poin
		}
	}
	if dijawabN == 0 {
		return nil, 0, 0, "tidak ada soal", true
	}
	if scored == 0 || penyebut <= 0 {
		// Semua yang dijawab belum dikoreksi -> belum ada nilai.
		return nil, dijawabN, belumDikoreksi, "belum dikoreksi", false
	}
	v := round2(sumPoin / penyebut * 100)
	if belumDikoreksi > 0 {
		return &v, dijawabN, belumDikoreksi, "sebagian belum dikoreksi", false
	}
	return &v, dijawabN, 0, "", false
}

func avgFloat(vals []float64) *float64 {
	if len(vals) == 0 {
		return nil
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	r := round2(sum / float64(len(vals)))
	return &r
}

// computeKompetensi menghitung nilai per kompetensi per siswa + agregat kelas.
func computeKompetensi(jadwalID string, soals []soalInput, attempts []attemptInput, jawabans []jawabanInput, siswaByAttempt map[string]siswaInput) kompetensiResult {
	jawabanPerSoal := map[string]map[string]jawabanInput{}
	for _, j := range jawabans {
		if jawabanPerSoal[j.SoalID] == nil {
			jawabanPerSoal[j.SoalID] = map[string]jawabanInput{}
		}
		jawabanPerSoal[j.SoalID][j.AttemptID] = j
	}
	// Kelompok soal per kompetensi (soal tanpa kompetensi -> grup "Tanpa Kompetensi").
	type grupKey struct {
		id   string // "" untuk tanpa kompetensi
		kode string
		nama string
	}
	grupOrder := []grupKey{}
	grupSoal := map[string][]soalInput{}
	keyOf := func(s soalInput) grupKey {
		if s.KompetensiID == nil {
			return grupKey{"", "-", "Tanpa Kompetensi"}
		}
		return grupKey{*s.KompetensiID, s.KompetensiKode, s.KompetensiNama}
	}
	for _, s := range soals {
		k := keyOf(s)
		if _, ok := grupSoal[k.id]; !ok {
			grupOrder = append(grupOrder, k)
		}
		grupSoal[k.id] = append(grupSoal[k.id], s)
	}
	// Urutkan grup berdasarkan kode (grup tanpa kompetensi terakhir).
	sort.SliceStable(grupOrder, func(i, j int) bool {
		if grupOrder[i].kode == "-" && grupOrder[j].kode != "-" {
			return false
		}
		if grupOrder[i].kode != "-" && grupOrder[j].kode == "-" {
			return true
		}
		return grupOrder[i].kode < grupOrder[j].kode
	})

	siswaOut := make([]kompetensiSiswa, 0, len(attempts))
	perKompetensiVals := map[string][]float64{}
	var rataRataVals []float64
	for _, a := range attempts {
		perAttempt := map[string]jawabanInput{}
		for soalID, m := range jawabanPerSoal {
			if j, ok := m[a.ID]; ok {
				perAttempt[soalID] = j
			}
		}
		info := siswaByAttempt[a.ID]
		row := kompetensiSiswa{
			SiswaID: a.SiswaID, NISN: info.NISN, Nama: info.Nama, Kelas: info.Kelas,
		}
		for _, g := range grupOrder {
			nilai, dijawabN, belumN, flag, tidakAda := nilaiKompetensi(grupSoal[g.id], perAttempt, a.ID)
			var gid *string
			if g.id != "" {
				v := g.id
				gid = &v
			}
			row.NilaiPerKompetensi = append(row.NilaiPerKompetensi, kompetensiNilai{
				KompetensiID: gid, KompetensiKode: g.kode, Kompetensi: g.nama,
				Nilai: nilai, JumlahSoal: len(grupSoal[g.id]),
				Dijawab: dijawabN, BelumDikoreksi: belumN,
				TidakAdaSoal: tidakAda, Flag: flag,
			})
			if nilai != nil {
				perKompetensiVals[g.id] = append(perKompetensiVals[g.id], *nilai)
			}
		}
		rata, _, _, _, _ := nilaiKompetensi(soals, perAttempt, a.ID)
		row.RataRata = rata
		if rata != nil {
			rataRataVals = append(rataRataVals, *rata)
		}
		siswaOut = append(siswaOut, row)
	}
	sort.SliceStable(siswaOut, func(i, j int) bool {
		if siswaOut[i].Nama != siswaOut[j].Nama {
			return siswaOut[i].Nama < siswaOut[j].Nama
		}
		return siswaOut[i].SiswaID < siswaOut[j].SiswaID
	})

	agregat := make([]kompetensiKelas, 0, len(grupOrder))
	for _, g := range grupOrder {
		vals := perKompetensiVals[g.id]
		var gid *string
		if g.id != "" {
			v := g.id
			gid = &v
		}
		agregat = append(agregat, kompetensiKelas{
			KompetensiID: gid, KompetensiKode: g.kode, Kompetensi: g.nama,
			RataRata: avgFloat(vals), JumlahSiswa: len(vals),
		})
	}

	return kompetensiResult{
		JadwalID: jadwalID, JumlahSiswa: len(attempts),
		RataRataKelas:         avgFloat(rataRataVals),
		RataRataPerKompetensi: agregat, Siswa: siswaOut,
	}
}

// ---------------------------------------------------------------------------
// Pemuatan data dari DB.
// ---------------------------------------------------------------------------

// selesaiFilter: attempt selesai = status 'selesai' ATAU punya nilai_akhir final.
const selesaiFilter = `(a.status = 'selesai' OR EXISTS (SELECT 1 FROM nilai_akhir na2 WHERE na2.attempt_id = a.id AND na2.status = 'final'))`

func loadDasar(ctx context.Context, pool *pgxpool.Pool, jadwalID string) (soals []soalInput, attempts []attemptInput, jawabans []jawabanInput, err error) {
	var paketID string
	err = pool.QueryRow(ctx,
		`SELECT paket_id FROM jadwal_ujian WHERE id = $1`, jadwalID).Scan(&paketID)
	if err != nil {
		return nil, nil, nil, errNotFound
	}

	rows, err := pool.Query(ctx,
		`SELECT s.id, COALESCE(ps.urutan, 0), s.pertanyaan, s.tipe, s.bobot::float8,
		        s.kompetensi_id, COALESCE(k.kode, ''), COALESCE(k.deskripsi, ''), s.jawaban
		 FROM paket_soal ps
		 JOIN soal s ON s.id = ps.soal_id
		 LEFT JOIN kompetensi k ON k.id = s.kompetensi_id
		 WHERE ps.paket_id = $1
		 ORDER BY ps.urutan ASC, s.pertanyaan ASC`, paketID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var s soalInput
		var jawabanRaw json.RawMessage
		var kompetensiID *string
		if err := rows.Scan(&s.ID, &s.Urutan, &s.Pertanyaan, &s.Tipe, &s.Bobot,
			&kompetensiID, &s.KompetensiKode, &s.KompetensiNama, &jawabanRaw); err != nil {
			return nil, nil, nil, err
		}
		s.KompetensiID = kompetensiID
		s.Opsi = parseOpsiSoal(jawabanRaw)
		soals = append(soals, s)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, err
	}

	rows, err = pool.Query(ctx,
		`SELECT a.id, a.siswa_id, na.nilai
		 FROM attempt a LEFT JOIN nilai_akhir na ON na.attempt_id = a.id
		 WHERE a.jadwal_id = $1 AND `+selesaiFilter+`
		 ORDER BY a.id`, jadwalID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var a attemptInput
		var nilai *float64
		if err := rows.Scan(&a.ID, &a.SiswaID, &nilai); err != nil {
			return nil, nil, nil, err
		}
		a.Nilai = nilai
		attempts = append(attempts, a)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, err
	}

	rows, err = pool.Query(ctx,
		`SELECT j.attempt_id, j.soal_id, j.jawaban, j.poin::float8, j.dikoreksi
		 FROM jawaban j JOIN attempt a ON a.id = j.attempt_id
		 WHERE a.jadwal_id = $1 AND `+selesaiFilter, jadwalID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var j jawabanInput
		var poin *float64
		var jawabanRaw *string
		// jawaban JSONB: baca sebagai teks lalu jadikan RawMessage (NULL -> nil).
		if err := rows.Scan(&j.AttemptID, &j.SoalID, &jawabanRaw, &poin, &j.Dikoreksi); err != nil {
			return nil, nil, nil, err
		}
		if jawabanRaw != nil {
			j.Jawaban = json.RawMessage(*jawabanRaw)
		}
		j.Poin = poin
		jawabans = append(jawabans, j)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, err
	}
	return soals, attempts, jawabans, nil
}

var errNotFound = fmt.Errorf("jadwal tidak ditemukan")

// parseOpsiSoal mengambil {"opsi":[{"label","teks"}, ...]} dari soal.jawaban.
func parseOpsiSoal(raw json.RawMessage) []opsiSoal {
	var obj struct {
		Opsi []struct {
			Label string `json:"label"`
			Teks  string `json:"teks"`
		} `json:"opsi"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil
	}
	out := make([]opsiSoal, 0, len(obj.Opsi))
	for _, o := range obj.Opsi {
		out = append(out, opsiSoal{Label: o.Label, Teks: o.Teks})
	}
	return out
}

func loadSiswa(ctx context.Context, pool *pgxpool.Pool, jadwalID string) (map[string]siswaInput, error) {
	rows, err := pool.Query(ctx,
		`SELECT a.id, a.siswa_id, COALESCE(s.nisn, ''), COALESCE(s.nama, u.nama, ''),
		        COALESCE(k.nama, '')
		 FROM attempt a
		 LEFT JOIN siswa s ON s.id = a.siswa_id OR s.user_id = a.siswa_id
		 LEFT JOIN users u ON u.id = a.siswa_id
		 LEFT JOIN kelas k ON k.id = s.kelas_id
		 WHERE a.jadwal_id = $1 AND `+selesaiFilter, jadwalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]siswaInput{}
	for rows.Next() {
		var info siswaInput
		if err := rows.Scan(&info.AttemptID, &info.SiswaID, &info.NISN, &info.Nama, &info.Kelas); err != nil {
			return nil, err
		}
		if _, ada := out[info.AttemptID]; !ada {
			out[info.AttemptID] = info
		}
	}
	return out, rows.Err()
}

func validJadwalID(c *echo.Context) (string, error) {
	id := strings.TrimSpace(c.Param("jadwalId"))
	if _, err := uuid.Parse(id); err != nil {
		return "", writeError(c, http.StatusBadRequest, "validation_error", "ID jadwal tidak valid.")
	}
	return id, nil
}

// ---------------------------------------------------------------------------
// Register.
// ---------------------------------------------------------------------------

// Register mendaftarkan endpoint analisis butir soal + nilai per kompetensi.
//
//	GET /api/v1/analisis/:jadwalId
//	GET /api/v1/analisis/:jadwalId/export.csv
//	GET /api/v1/kompetensi/:jadwalId
//	GET /api/v1/kompetensi/:jadwalId/export.csv
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	// GET /api/v1/analisis/:jadwalId
	e.GET("/api/v1/analisis/:jadwalId", func(c *echo.Context) error {
		_, role, _, ok := requirePerm(c, "analisis.read")
		if !ok {
			if _, _, _, _, okUser := audit.CurrentUser(c); !okUser {
				return writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login.")
			}
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat analisis butir soal.")
		}
		jadwalID, err := validJadwalID(c)
		if err != nil {
			return err
		}
		soals, attempts, jawabans, err := loadDasar(c.Request().Context(), pool, jadwalID)
		if err != nil {
			if err == errNotFound {
				return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghitung analisis butir soal.")
		}
		// SoD: proktor melihat statistik TANPA teks pertanyaan.
		return writeOK(c, SembunyikanPertanyaanUntukProktor(computeAnalisis(jadwalID, soals, attempts, jawabans), role))
	})

	// GET /api/v1/analisis/:jadwalId/export.csv
	e.GET("/api/v1/analisis/:jadwalId/export.csv", func(c *echo.Context) error {
		userID, role, _, ok := requirePerm(c, "analisis.export")
		if !ok {
			if _, _, _, _, okUser := audit.CurrentUser(c); !okUser {
				return writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login.")
			}
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin export analisis butir soal.")
		}
		jadwalID, err := validJadwalID(c)
		if err != nil {
			return err
		}
		soals, attempts, jawabans, err := loadDasar(c.Request().Context(), pool, jadwalID)
		if err != nil {
			if err == errNotFound {
				return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghitung analisis butir soal.")
		}
		// SoD: proktor mengexport statistik TANPA teks pertanyaan
		// (tanpa ini, proktor bisa mengelak dari penyembunyian di
		// endpoint JSON lewat export CSV).
		hasil := SembunyikanPertanyaanUntukProktor(computeAnalisis(jadwalID, soals, attempts, jawabans), role)
		_ = audit.Catat(c.Request().Context(), pool, userID, "analisis.export", "analisis", jadwalID,
			map[string]any{"jadwal_id": jadwalID, "jumlah_soal": hasil.JumlahSoal, "jumlah_peserta": hasil.JumlahPeserta}, c.RealIP())

		filename := fmt.Sprintf("analisis-butir-%s-%s.csv", jadwalID, time.Now().Format("2006-01-02"))
		c.Response().Header().Set("Content-Type", "text/csv; charset=utf-8")
		c.Response().Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
		w := csv.NewWriter(c.Response())
		_, _ = c.Response().Write([]byte{0xEF, 0xBB, 0xBF}) // BOM UTF-8 agar Excel benar
		_ = w.Write([]string{"no", "soal_id", "pertanyaan", "tipe", "bobot", "kompetensi",
			"jumlah_penjab", "jumlah_benar", "kesukaran", "label_kesukaran",
			"daya_beda", "interpretasi_daya_beda", "opsi", "pemilih", "persen", "pengecoh_berfungsi"})
		for i, item := range hasil.Items {
			kesukaran := ""
			if item.Kesukaran != nil {
				kesukaran = fmt.Sprintf("%.2f", *item.Kesukaran)
			}
			dayaBeda := ""
			if item.DayaBeda != nil {
				dayaBeda = fmt.Sprintf("%.2f", *item.DayaBeda)
			}
			if item.PengecohStatus == "n/a" || len(item.Pengecoh) == 0 {
				_ = w.Write([]string{
					fmt.Sprint(i + 1), item.SoalID, item.Pertanyaan, item.Tipe,
					fmt.Sprintf("%g", item.Bobot), item.KompetensiKode,
					fmt.Sprint(item.JumlahPenjab), fmt.Sprint(item.JumlahBenar),
					kesukaran, item.LabelKesukaran, dayaBeda, item.InterpretasiDayaBeda,
					"N/A", "", "", "",
				})
				continue
			}
			for _, p := range item.Pengecoh {
				berfungsi := "tidak berfungsi"
				if p.Berfungsi {
					berfungsi = "berfungsi"
				}
				_ = w.Write([]string{
					fmt.Sprint(i + 1), item.SoalID, item.Pertanyaan, item.Tipe,
					fmt.Sprintf("%g", item.Bobot), item.KompetensiKode,
					fmt.Sprint(item.JumlahPenjab), fmt.Sprint(item.JumlahBenar),
					kesukaran, item.LabelKesukaran, dayaBeda, item.InterpretasiDayaBeda,
					p.Opsi, fmt.Sprint(p.Pemilih), fmt.Sprintf("%.2f", p.Persen), berfungsi,
				})
			}
		}
		w.Flush()
		return nil
	})

	// GET /api/v1/kompetensi/:jadwalId
	e.GET("/api/v1/kompetensi/:jadwalId", func(c *echo.Context) error {
		_, role, siswaID, ok := requirePerm(c, "rekap.read")
		if !ok {
			if _, _, _, _, okUser := audit.CurrentUser(c); !okUser {
				return writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login.")
			}
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat nilai per kompetensi.")
		}
		jadwalID, err := validJadwalID(c)
		if err != nil {
			return err
		}
		ctx := c.Request().Context()
		soals, attempts, jawabans, err := loadDasar(ctx, pool, jadwalID)
		if err != nil {
			if err == errNotFound {
				return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghitung nilai per kompetensi.")
		}
		// Siswa hanya boleh melihat nilainya sendiri (anti-IDOR, KONTRAK §5.3).
		if role == "siswa" {
			filtered := attempts[:0]
			for _, a := range attempts {
				if a.SiswaID == siswaID {
					filtered = append(filtered, a)
				}
			}
			attempts = filtered
		}
		siswaByAttempt, err := loadSiswa(ctx, pool, jadwalID)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca data siswa.")
		}
		return writeOK(c, computeKompetensi(jadwalID, soals, attempts, jawabans, siswaByAttempt))
	})

	// GET /api/v1/kompetensi/:jadwalId/export.csv
	e.GET("/api/v1/kompetensi/:jadwalId/export.csv", func(c *echo.Context) error {
		userID, role, siswaID, ok := requirePerm(c, "rekap.export")
		if !ok {
			if _, _, _, _, okUser := audit.CurrentUser(c); !okUser {
				return writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login.")
			}
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin export nilai per kompetensi.")
		}
		jadwalID, err := validJadwalID(c)
		if err != nil {
			return err
		}
		ctx := c.Request().Context()
		soals, attempts, jawabans, err := loadDasar(ctx, pool, jadwalID)
		if err != nil {
			if err == errNotFound {
				return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghitung nilai per kompetensi.")
		}
		if role == "siswa" {
			filtered := attempts[:0]
			for _, a := range attempts {
				if a.SiswaID == siswaID {
					filtered = append(filtered, a)
				}
			}
			attempts = filtered
		}
		siswaByAttempt, err := loadSiswa(ctx, pool, jadwalID)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca data siswa.")
		}
		hasil := computeKompetensi(jadwalID, soals, attempts, jawabans, siswaByAttempt)
		_ = audit.Catat(ctx, pool, userID, "rekap.export", "kompetensi", jadwalID,
			map[string]any{"jadwal_id": jadwalID, "jumlah_siswa": hasil.JumlahSiswa}, c.RealIP())

		filename := fmt.Sprintf("nilai-kompetensi-%s-%s.csv", jadwalID, time.Now().Format("2006-01-02"))
		c.Response().Header().Set("Content-Type", "text/csv; charset=utf-8")
		c.Response().Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
		w := csv.NewWriter(c.Response())
		_, _ = c.Response().Write([]byte{0xEF, 0xBB, 0xBF})
		_ = w.Write([]string{"nisn", "nama", "kelas", "kompetensi_kode", "kompetensi",
			"nilai", "jumlah_soal", "dijawab", "flag", "rata_rata_siswa"})
		for _, s := range hasil.Siswa {
			rata := ""
			if s.RataRata != nil {
				rata = fmt.Sprintf("%.2f", *s.RataRata)
			}
			for _, k := range s.NilaiPerKompetensi {
				nilai := ""
				if k.Nilai != nil {
					nilai = fmt.Sprintf("%.2f", *k.Nilai)
				}
				_ = w.Write([]string{
					s.NISN, s.Nama, s.Kelas, k.KompetensiKode, k.Kompetensi,
					nilai, fmt.Sprint(k.JumlahSoal), fmt.Sprint(k.Dijawab), k.Flag, rata,
				})
			}
		}
		w.Flush()
		return nil
	})
}
