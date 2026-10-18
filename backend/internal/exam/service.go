package exam

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"hash/fnv"
	mrand "math/rand"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"cbt-smk-v2/backend/internal/platform"
	"cbt-smk-v2/backend/internal/grading"
	"cbt-smk-v2/backend/internal/monitor"
)

// --- Service ----------------------------------------------------------------


// Service adalah lapisan logika bisnis paket exam.
// repository adalah kontrak penyimpanan yang dipakai Service.
// (DIP: Service bergantung pada abstraksi, bukan *Repository konkrit.)
type repository interface {
	AktifkanJadwal(ctx context.Context, q querier, id string) error
	AttemptMonitorCounts(ctx context.Context, attemptID string) (dijawab, ragu, pelanggaran int)
	AttemptRingkasan(ctx context.Context, attemptID, paketID string) (dijawab, ragu, total int)
	AttemptStatusJadwal(ctx context.Context, q querier, attemptID string) (status, jadwalID string)
	BekukanTimer(ctx context.Context, q querier, id string) error
	BlokirAttempt(ctx context.Context, attemptID string) error
	CatatAudit(ctx context.Context, actorID, aksi, entitas, entitasID string, detail any, ip string)
	CountAttemptJadwal(ctx context.Context, id string) (int, error)
	CountJadwalPaket(ctx context.Context, paketID string) (int, error)
	CountSoalMapelLain(ctx context.Context, paketID, mapelID string) (int, error)
	CountUsulanMenunggu(ctx context.Context, attemptID string) (int, error)
	DeleteJadwal(ctx context.Context, id string) (int64, error)
	DeletePaket(ctx context.Context, paketID string) (int64, error)
	DeletePaketSoal(ctx context.Context, paketID, soalID string) (int64, error)
	DraftKeAktif(ctx context.Context, q querier, jadwalID string) error
	FindAttemptByStatus(ctx context.Context, jadwalID, siswaID, status string) (attempt, error)
	GetAttempt(ctx context.Context, id string) (attempt, error)
	GetJadwal(ctx context.Context, id string) (jadwal, error)
	GetJawabanKoreksi(ctx context.Context, jawabanID string) (attemptID, soalID string, bobot float64, err error)
	GetPaketDetail(ctx context.Context, id string) (paket, error)
	InsertAttempt(ctx context.Context, q querier, attemptID, jadwalID, siswaID string, selesai time.Time, urutanSoalJSON, urutanOpsiJSON []byte) error
	InsertJadwal(ctx context.Context, j jadwal) (string, error)
	InsertPaket(ctx context.Context, p paket) (string, error)
	InsertPaketSoal(ctx context.Context, paketID, soalID string, urutan int) error
	InsertPelanggaran(ctx context.Context, attemptID, jenis, detail string) error
	InsertUsulan(ctx context.Context, attemptID, pengawasID, alasan string) (string, error)
	InsertUsulanDisetujui(ctx context.Context, q querier, attemptID, uid, alasan string)
	JadwalStatus(ctx context.Context, jadwalID string) string
	JadwalStatusOnly(ctx context.Context, id string) (string, error)
	KelasExists(ctx context.Context, kelasID string) (bool, error)
	KelasSiswa(ctx context.Context, siswaID string) (string, bool)
	LanjutkanAttempt(ctx context.Context, q querier, attemptID string, sisa int) error
	ListHasilJawaban(ctx context.Context, attemptID string) ([]hasilJawabanRow, error)
	ListJadwal(ctx context.Context, status, kelasID, tanggal string) ([]jadwalItem, error)
	ListJawabanNilai(ctx context.Context, attemptID string) ([]jawabanNilaiRow, error)
	ListJawabanTersimpan(ctx context.Context, attemptID string) ([]jawabanTersimpanRow, error)
	ListKoreksi(ctx context.Context, jadwalID string) ([]koreksiItem, error)
	ListPaket(ctx context.Context, jenis, mapelID string) ([]paket, error)
	ListPaketSoal(ctx context.Context, paketID string) ([]paketSoalItem, error)
	ListSoalAcak(ctx context.Context, mapelID, paketID string, jumlah int) ([]string, error)
	ListSoalDetail(ctx context.Context, paketID string) ([]soalDetailRow, error)
	ListSoalHasilPaket(ctx context.Context, paketID string) ([]soalHasilRow, error)
	ListSoalNilaiKosong(ctx context.Context, attemptID string) ([]soalNilaiRow, error)
	ListSoalPaket(ctx context.Context, paketID string) ([]soalPaketRow, error)
	ListUjianAktif(ctx context.Context, siswaID, kelasID string) ([]ujianAktifItem, error)
	ListUsulanUnblock(ctx context.Context, jadwalID string) ([]usulanItem, error)
	MapelExists(ctx context.Context, mapelID string) (bool, error)
	MaxPelanggaran(ctx context.Context) int
	MaxUrutanPaketSoal(ctx context.Context, paketID string) (int, error)
	PaketExists(ctx context.Context, paketID string) (bool, error)
	PaketMapelID(ctx context.Context, paketID string) (string, error)
	PaketSoalExists(ctx context.Context, paketID, soalID string) (bool, error)
	PauseJadwalStatus(ctx context.Context, q querier, id string) error
	PelanggaranTerakhir(ctx context.Context, attemptID string) (*time.Time, int)
	PulihkanTimer(ctx context.Context, q querier, id string) error
	RaguLama(ctx context.Context, attemptID, soalID string) bool
	SetujuiUsulan(ctx context.Context, q querier, usulanID, uid string)
	SinkronJumlahSoal(ctx context.Context, paketID string) error
	SiswaIDByUser(ctx context.Context, userID string) (string, error)
	SoalMapelID(ctx context.Context, soalID string) (string, error)
	SoalUntukSkor(ctx context.Context, soalID string) (tipe string, kunci, bobotRaw, tampilan []byte, err error)
	TampilHasilEfektif(ctx context.Context, paketTampil bool) bool
	TandaiSelesai(ctx context.Context, attemptID string, nilai float64) error
	TokenExpiryDetik(ctx context.Context) int
	ToleransiTelatMenit(ctx context.Context) int
	UpdateJadwal(ctx context.Context, id string, j jadwal) (int64, error)
	UpdateJawabanPoin(ctx context.Context, q querier, jawabanID string, poin float64) error
	UpdateNilaiAkhir(ctx context.Context, attemptID string, nilai float64) error
	UpdatePaket(ctx context.Context, id string, p paket) (int64, error)
	UpdateToken(ctx context.Context, id, hash string, expiry time.Time) error
	UpsertJawaban(ctx context.Context, attemptID, soalID string, jawaban []byte, ragu bool, poinArg any, dikoreksi bool) error
	UpsertNilaiAkhir(ctx context.Context, attemptID, jadwalID, siswaID string, nilai float64) error
	UpsertRagu(ctx context.Context, attemptID, soalID string, ragu bool) error
	UsulanMenungguTerbaru(ctx context.Context, q querier, attemptID string) (string, error)
}

type Service struct {
	repo repository
	pool *pgxpool.Pool
}

// NewService membuat Service.
func NewService(repo repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, pool: pool}
}

// catat menulis audit log. (Dulu: auditCatat di exam.go.)
func (s *Service) catat(ctx context.Context, actor platform.RequestActor, aksi, entitas, entitasID string, detail any) {
	s.repo.CatatAudit(ctx, actor.UserID, aksi, entitas, entitasID, detail, actor.RealIP)
}

// --- Fungsi murni -----------------------------------------------------------
// (Dipindah dari exam.go tanpa perubahan; diuji langsung oleh
// exam_hardening_test.go — nama + signature dipertahankan.)

// jakarta memuat zona Asia/Jakarta (fallback: Local bila gagal).
func jakarta() *time.Location {
	if loc, err := time.LoadLocation("Asia/Jakarta"); err == nil {
		return loc
	}
	return time.Local
}

// gabungTanggalJam menggabungkan tanggal "2006-01-02" + jam menjadi
// time.Time di zona sekolah (Asia/Jakarta). Jam diterima dalam format
// "15:04:05" maupun "15:04" (Fase 3: "HH:MM" dinormalisasi ke
// "HH:MM:SS" agar POST jadwal dengan jam "08:00" lolos). Fraksi detik
// pada jam (mis. "15:04:05.123") dipotong agar Parse tidak gagal.
func gabungTanggalJam(tanggal, jam string) (time.Time, error) {
	if i := strings.Index(jam, "."); i >= 0 {
		jam = jam[:i]
	}
	jam = strings.TrimSpace(jam)
	if len(jam) == 5 && jam[2] == ':' {
		jam += ":00"
	}
	return time.ParseInLocation("2006-01-02 15:04:05", tanggal+" "+jam, jakarta())
}

func newUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // versi 4
	b[8] = (b[8] & 0x3f) | 0x80 // varian
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// sisaWaktu menghitung sisa waktu (detik) attempt dari sisi server.
// Bila jadwal dijeda -> sisa beku. Bila deadline lewat -> 0.
func sisaWaktu(a attempt, jadwalStatus string, now time.Time) int {
	if jadwalStatus == "pause" && a.SisaWaktuBeku != nil {
		if *a.SisaWaktuBeku < 0 {
			return 0
		}
		return *a.SisaWaktuBeku
	}
	if dl, ok := a.Deadline.(time.Time); ok {
		sisa := int(dl.Sub(now).Seconds())
		if sisa < 0 {
			return 0
		}
		return sisa
	}
	return 0
}

// shuffleSeed mengocok salinan slice dengan seed deterministik per attempt.
func shuffleSeed(items []string, seed int64) []string {
	out := append([]string(nil), items...)
	r := mrand.New(mrand.NewSource(seed))
	r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

func seedDariAttemptID(id string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(id))
	return int64(h.Sum64())
}

// shuffleTampilan mengocok data tampil soal Tim 3 (kolom soal.jawaban)
// bila paket acak_jawaban aktif:
//   - PG / PG Kompleks / PG Multi / Pilihan Gambar / Studi Kasus:
//     kocok array jawaban.opsi.
//   - Menjodohkan / Peta Konsep: kocok nilai "kanan" antar pasangan
//     (urutan "kiri" tetap), agar siswa tidak menghafal posisi.
//   - Tipe lain: tidak dikocok.
func shuffleTampilan(tampilanRaw []byte, tipe string, seed int64) json.RawMessage {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(tampilanRaw, &obj); err != nil {
		return tampilanRaw
	}
	kanon := grading.NormalisasiTipe(tipe)
	r := mrand.New(mrand.NewSource(seed))
	switch kanon {
	case "pg", "pg_kompleks":
		var opsi []json.RawMessage
		if err := json.Unmarshal(obj["opsi"], &opsi); err != nil || len(opsi) < 2 {
			return tampilanRaw
		}
		r.Shuffle(len(opsi), func(i, j int) { opsi[i], opsi[j] = opsi[j], opsi[i] })
		obj["opsi"], _ = json.Marshal(opsi)
	case "menjodohkan":
		var pasangan []map[string]json.RawMessage
		if err := json.Unmarshal(obj["pasangan"], &pasangan); err != nil || len(pasangan) < 2 {
			return tampilanRaw
		}
		kanan := make([]json.RawMessage, 0, len(pasangan))
		for _, p := range pasangan {
			kanan = append(kanan, p["kanan"])
		}
		r.Shuffle(len(kanan), func(i, j int) { kanan[i], kanan[j] = kanan[j], kanan[i] })
		for i := range pasangan {
			pasangan[i]["kanan"] = kanan[i]
		}
		obj["pasangan"], _ = json.Marshal(pasangan)
	default:
		return tampilanRaw
	}
	out, err := json.Marshal(obj)
	if err != nil {
		return tampilanRaw
	}
	return out
}

// kunciUntukGrading memperkaya kunci untuk tipe mengurutkan: grading
// butuh daftar item asli (dari tampilan soal) untuk membandingkan
// jawaban string siswa. Frontend mengirim jawaban indeks, jadi ini
// hanya cadangan bila jawaban berupa string item.
func kunciUntukGrading(kunci, tampilan []byte, tipe string) []byte {
	if grading.NormalisasiTipe(tipe) != "mengurutkan" {
		return kunci
	}
	var kunciObj map[string]json.RawMessage
	if err := json.Unmarshal(kunci, &kunciObj); err != nil {
		kunciObj = map[string]json.RawMessage{}
	}
	var tampilanObj map[string]json.RawMessage
	if err := json.Unmarshal(tampilan, &tampilanObj); err == nil {
		if item, ada := tampilanObj["item"]; ada {
			kunciObj["item"] = item
		}
	}
	out, err := json.Marshal(kunciObj)
	if err != nil {
		return kunci
	}
	return out
}

// kunciTampil merangkum kunci benar untuk halaman hasil (Tryout /
// tampil_hasil). Hanya dipakai SETELAH ujian selesai, tidak pernah
// sebelum/selama ujian. Mengembalikan teks ringkas, bukan kunci mentah.
func kunciTampil(tipe string, kunci, tampilan []byte) string {
	kanon := grading.NormalisasiTipe(tipe)
	switch kanon {
	case "pg":
		if opsi, ok := func() ([]string, bool) {
			var obj map[string]json.RawMessage
			if err := json.Unmarshal(kunci, &obj); err == nil {
				if raw, ada := obj["opsi"]; ada {
					var s string
					if err := json.Unmarshal(raw, &s); err == nil {
						return []string{s}, true
					}
				}
			}
			return nil, false
		}(); ok && len(opsi) > 0 {
			// Cari teks opsi dari tampilan.
			var tamp map[string]json.RawMessage
			if err := json.Unmarshal(tampilan, &tamp); err == nil {
				var opsiList []struct {
					Label string `json:"label"`
					Teks  string `json:"teks"`
				}
				if err := json.Unmarshal(tamp["opsi"], &opsiList); err == nil {
					for _, o := range opsiList {
						if o.Label == opsi[0] {
							return "Jawaban benar: " + o.Label + " — " + o.Teks
						}
					}
				}
			}
			return "Jawaban benar: " + opsi[0]
		}
		return ""
	case "pg_kompleks":
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(kunci, &obj); err == nil {
			var arr []string
			if err := json.Unmarshal(obj["opsi"], &arr); err == nil {
				return "Jawaban benar: " + strings.Join(arr, ", ")
			}
		}
		return ""
	case "benar_salah":
		var obj map[string]json.RawMessage
		var nilai []bool
		if err := json.Unmarshal(kunci, &obj); err == nil {
			_ = json.Unmarshal(obj["nilai"], &nilai)
		}
		var tamp map[string]json.RawMessage
		var pernyataan []string
		if err := json.Unmarshal(tampilan, &tamp); err == nil {
			_ = json.Unmarshal(tamp["pernyataan"], &pernyataan)
		}
		parts := []string{}
		for i, v := range nilai {
			label := ""
			if i < len(pernyataan) {
				label = pernyataan[i]
			}
			bs := "Salah"
			if v {
				bs = "Benar"
			}
			parts = append(parts, label+" = "+bs)
		}
		return strings.Join(parts, "; ")
	case "menjodohkan":
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(kunci, &obj); err == nil {
			var pasangan []struct {
				Kiri  string `json:"kiri"`
				Kanan string `json:"kanan"`
			}
			if err := json.Unmarshal(obj["pasangan"], &pasangan); err == nil {
				parts := []string{}
				for _, p := range pasangan {
					parts = append(parts, p.Kiri+" → "+p.Kanan)
				}
				return "Pasangan benar: " + strings.Join(parts, "; ")
			}
		}
		return ""
	case "isian_singkat", "isi_ganda", "teks_rumpang":
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(kunci, &obj); err == nil {
			var arr []string
			if err := json.Unmarshal(obj["jawaban"], &arr); err == nil {
				return "Jawaban yang diterima: " + strings.Join(arr, " | ")
			}
		}
		return ""
	case "mengurutkan":
		var obj map[string]json.RawMessage
		var urutan []int
		if err := json.Unmarshal(kunci, &obj); err == nil {
			var nums []float64
			if err := json.Unmarshal(obj["urutan"], &nums); err == nil {
				for _, n := range nums {
					urutan = append(urutan, int(n))
				}
			}
		}
		var tamp map[string]json.RawMessage
		var item []string
		if err := json.Unmarshal(tampilan, &tamp); err == nil {
			_ = json.Unmarshal(tamp["item"], &item)
		}
		if len(urutan) > 0 && len(item) > 0 {
			parts := []string{}
			for _, u := range urutan {
				if u >= 0 && u < len(item) {
					parts = append(parts, item[u])
				}
			}
			return "Urutan benar: " + strings.Join(parts, " → ")
		}
		return ""
	case "uraian":
		return "Dinilai manual oleh guru (rubrik)."
	case "skala_sikap":
		return "Tidak dinilai (survei)."
	default:
		return ""
	}
}

// --- Nilai & lifecycle attempt ---------------------------------------------

// hitungNilai menghitung nilai akhir attempt dari tabel jawaban + soal.
// Dipanggil saat submit & saat koreksi (attempt selesai).
func (s *Service) hitungNilai(ctx context.Context, attemptID string) (float64, bool, error) {
	rows, err := s.repo.ListJawabanNilai(ctx, attemptID)
	if err != nil {
		return 0, false, err
	}
	var list []grading.JawabanNilai
	for _, j := range rows {
		kanon := grading.NormalisasiTipe(j.Tipe)
		dinilai := grading.Dinilai(j.Tipe, nil)
		menunggu := false
		poin := j.Poin
		if kanon == "uraian" && !j.Dikoreksi {
			// Soal manual belum dikoreksi -> poin 0 sementara, menunggu.
			menunggu = true
			poin = 0
		}
		if !j.AdaJawaban {
			poin = 0
		}
		list = append(list, grading.JawabanNilai{
			Poin: poin, Bobot: j.Bobot, MenungguKoreksi: menunggu, Dinilai: dinilai,
		})
	}
	// Soal dalam paket yang BELUM punya baris jawaban = kosong (poin 0).
	// Hitung lewat paket agar penyebut lengkap.
	kosong, err := s.repo.ListSoalNilaiKosong(ctx, attemptID)
	if err != nil {
		return 0, false, err
	}
	for _, k := range kosong {
		if grading.Dinilai(k.Tipe, nil) {
			list = append(list, grading.JawabanNilai{Poin: 0, Bobot: k.Bobot, Dinilai: true})
		}
	}
	nilai, menunggu := grading.HitungNilaiAkhir(list)
	return nilai, menunggu, nil
}

// syncNilaiAkhir menyalin nilai attempt ke tabel nilai_akhir (Tim 5) agar
// endpoint rekap membaca nilai yang sama dengan yang dihitung mesin
// ujian. Dipanggil dari finalisasiAttempt (submit/auto-submit) dan
// koreksiJawaban. status dibiarkan 'menunggu_koreksi' sampai proktor
// menjalankan finalisasi rekap (yang mengubahnya jadi 'final') —
// sync TIDAK PERNAH menurunkan status 'final' yang sudah ada.
// Best-effort: gagal sync tidak menggagalkan alur ujian.
func (s *Service) syncNilaiAkhir(ctx context.Context, a attempt, nilai float64) {
	// nilai_akhir.siswa_id merujuk siswa.id, sedangkan attempt.siswa_id
	// menyimpan users.id (auth_siswa_id) — petakan lewat siswa.user_id.
	siswaID, err := s.repo.SiswaIDByUser(ctx, a.SiswaID)
	if err != nil {
		return
	}
	_ = s.repo.UpsertNilaiAkhir(ctx, a.ID, a.JadwalID, siswaID, nilai)
}

// finalisasiAttempt menandai attempt selesai + menghitung nilai akhir.
func (s *Service) finalisasiAttempt(ctx context.Context, a attempt) (float64, bool, error) {
	nilai, menunggu, err := s.hitungNilai(ctx, a.ID)
	if err != nil {
		return 0, false, err
	}
	if err := s.repo.TandaiSelesai(ctx, a.ID, nilai); err != nil {
		return 0, false, err
	}
	s.syncNilaiAkhir(ctx, a, nilai)
	return nilai, menunggu, nil
}

// autoSubmitJikaLewat: bila attempt berjalan & deadline lewat, finalisasi.
// Mengembalikan attempt terbaru + apakah baru saja disubmit otomatis.
// Bila jadwal dijeda (pause), timer BEKU -> JANGAN auto-submit walau
// deadline (yang belum disesuaikan) sudah lewat.
func (s *Service) autoSubmitJikaLewat(ctx context.Context, a attempt, jadwalStatus string) (attempt, bool) {
	if a.Status != "berjalan" || jadwalStatus == "pause" {
		return a, false
	}
	dl, ok := a.Deadline.(time.Time)
	if !ok || time.Now().Before(dl) {
		return a, false
	}
	if _, _, err := s.finalisasiAttempt(ctx, a); err != nil {
		return a, false
	}
	baru, err := s.repo.GetAttempt(ctx, a.ID)
	if err != nil {
		return a, false
	}
	return baru, true
}

// broadcastAttempt mengirim update ringan monitor untuk satu attempt.
func (s *Service) broadcastAttempt(ctx context.Context, a attempt) {
	dijawab, ragu, pelanggaran := s.repo.AttemptMonitorCounts(ctx, a.ID)
	monitor.Broadcast(a.JadwalID, map[string]any{
		"type": "update", "jadwal_id": a.JadwalID, "attempt_id": a.ID,
		"status": a.Status, "dijawab": dijawab, "ragu": ragu,
		"pelanggaran": pelanggaran, "at": time.Now().Format(time.RFC3339),
	})
}
