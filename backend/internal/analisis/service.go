package analisis

import (
	"context"
	"encoding/json"
	"math"
	"sort"
	"strings"
	"unicode/utf8"

	"cbt-smk-v2/backend/internal/grading"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ---------------------------------------------------------------------------
// Input mentah (dipakai fungsi murni di bawah, agar bisa diuji tanpa DB).
// (Dipindah dari analisis.go tanpa perubahan.)
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
// (Dipindah dari analisis.go tanpa perubahan.)
// ---------------------------------------------------------------------------

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
// Service
// ---------------------------------------------------------------------------


// Service adalah lapisan logika bisnis paket analisis.
// repository adalah kontrak penyimpanan yang dipakai Service.
// (DIP: Service bergantung pada abstraksi, bukan *Repository konkrit.)
type repository interface {
	LoadDasar(ctx context.Context, jadwalID string) (soals []soalInput, attempts []attemptInput, jawabans []jawabanInput, err error)
	LoadSiswa(ctx context.Context, jadwalID string) (map[string]siswaInput, error)
}

type Service struct {
	repo repository
	pool *pgxpool.Pool
}

// NewService membuat Service.
func NewService(repo repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, pool: pool}
}


// muat memuat data dasar; memetakan error ke AppError sesuai endpoint.
func (s *Service) muat(ctx context.Context, jadwalID, pesanGagal string) ([]soalInput, []attemptInput, []jawabanInput, *platform.AppError) {
	soals, attempts, jawabans, err := s.repo.LoadDasar(ctx, jadwalID)
	if err != nil {
		if err == errNotFound {
			return nil, nil, nil, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
		}
		return nil, nil, nil, platform.ErrInternal("internal_error", pesanGagal)
	}
	return soals, attempts, jawabans, nil
}

// Analisis menghitung analisis butir soal satu jadwal.
// SoD: proktor melihat statistik TANPA teks pertanyaan.
func (s *Service) Analisis(ctx context.Context, jadwalID, role string) (analisisResult, *platform.AppError) {
	soals, attempts, jawabans, appErr := s.muat(ctx, jadwalID, "Gagal menghitung analisis butir soal.")
	if appErr != nil {
		return analisisResult{}, appErr
	}
	return SembunyikanPertanyaanUntukProktor(computeAnalisis(jadwalID, soals, attempts, jawabans), role), nil
}

// AnalisisExport seperti Analisis + mencatat audit export.
func (s *Service) AnalisisExport(ctx context.Context, actor platform.RequestActor, jadwalID, role string) (analisisResult, *platform.AppError) {
	hasil, appErr := s.Analisis(ctx, jadwalID, role)
	if appErr != nil {
		return analisisResult{}, appErr
	}
	platform.Catat(ctx, s.pool, actor, "analisis.export", "analisis", jadwalID,
		map[string]any{"jadwal_id": jadwalID, "jumlah_soal": hasil.JumlahSoal, "jumlah_peserta": hasil.JumlahPeserta})
	return hasil, nil
}

// filterSiswaUntukSiswa membatasi attempts ke milik siswa sendiri
// (anti-IDOR, KONTRAK §5.3). (Dulu: filter inline di handler.)
func filterSiswaUntukSiswa(attempts []attemptInput, role, siswaID string) []attemptInput {
	if role != "siswa" {
		return attempts
	}
	filtered := attempts[:0]
	for _, a := range attempts {
		if a.SiswaID == siswaID {
			filtered = append(filtered, a)
		}
	}
	return filtered
}

// Kompetensi menghitung nilai per kompetensi satu jadwal.
func (s *Service) Kompetensi(ctx context.Context, jadwalID, role, siswaID string) (kompetensiResult, *platform.AppError) {
	soals, attempts, jawabans, appErr := s.muat(ctx, jadwalID, "Gagal menghitung nilai per kompetensi.")
	if appErr != nil {
		return kompetensiResult{}, appErr
	}
	attempts = filterSiswaUntukSiswa(attempts, role, siswaID)
	siswaByAttempt, err := s.repo.LoadSiswa(ctx, jadwalID)
	if err != nil {
		return kompetensiResult{}, platform.ErrInternal("internal_error", "Gagal membaca data siswa.")
	}
	return computeKompetensi(jadwalID, soals, attempts, jawabans, siswaByAttempt), nil
}

// KompetensiExport seperti Kompetensi + mencatat audit export.
func (s *Service) KompetensiExport(ctx context.Context, actor platform.RequestActor, jadwalID, role, siswaID string) (kompetensiResult, *platform.AppError) {
	hasil, appErr := s.Kompetensi(ctx, jadwalID, role, siswaID)
	if appErr != nil {
		return kompetensiResult{}, appErr
	}
	platform.Catat(ctx, s.pool, actor, "rekap.export", "kompetensi", jadwalID,
		map[string]any{"jadwal_id": jadwalID, "jumlah_siswa": hasil.JumlahSiswa})
	return hasil, nil
}
