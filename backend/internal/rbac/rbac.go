// Package rbac mendefinisikan daftar permission kanonis dan matriks
// role -> permission untuk CBT SMK v2.
//
// ATURAN: daftar permission di sini adalah SATU-SATUNYA sumber kebenaran.
// Tim seed, middleware, dan dokumentasi (docs/KONTRAK.md) WAJIB memakai
// daftar yang sama persis. Jangan menambah permission baru tanpa
// memperbarui KONTRAK.md dan seed role_permissions.
package rbac

// Role adalah kode role yang stabil, disimpan di users.role.
type Role string

const (
	RoleAdmin      Role = "admin"
	RoleProktor    Role = "proktor"
	RoleGuru       Role = "guru"
	RolePengawas   Role = "pengawas"
	RoleSiswa      Role = "siswa"
	RolePengujiUKK Role = "penguji_ukk"
)

// AllRoles mengembalikan seluruh role yang dikenal sistem.
func AllRoles() []Role {
	return []Role{RoleAdmin, RoleProktor, RoleGuru, RolePengawas, RoleSiswa, RolePengujiUKK}
}

// AllPermissions adalah daftar permission LENGKAP dan kanonis.
// Jangan mengubah string di sini tanpa sinkronisasi ke KONTRAK.md,
// migrasi 0001 (seed permissions + role_permissions), dan tim seed.
var AllPermissions = []string{
	// Master data
	"master.jurusan.read", "master.jurusan.create", "master.jurusan.update", "master.jurusan.delete",
	"master.kelas.read", "master.kelas.create", "master.kelas.update", "master.kelas.delete",
	"master.mapel.read", "master.mapel.create", "master.mapel.update", "master.mapel.delete",
	"master.guru.read", "master.guru.create", "master.guru.update", "master.guru.delete",
	"master.siswa.read", "master.siswa.create", "master.siswa.update", "master.siswa.delete", "master.siswa.import",
	// Bank soal
	"banksoal.read", "banksoal.create", "banksoal.update", "banksoal.delete", "banksoal.lock", "banksoal.import",
	// Paket ujian
	"paket.read", "paket.create", "paket.update", "paket.delete",
	// Jadwal
	"jadwal.read", "jadwal.create", "jadwal.update", "jadwal.delete",
	"jadwal.pause", "jadwal.resume", "jadwal.unblock", "jadwal.usulan-unblock",
	// Ujian
	"ujian.kerjakan", "ujian.monitor",
	// Koreksi
	"koreksi.read", "koreksi.update",
	// Rekap
	"rekap.read", "rekap.export", "rekap.finalisasi", "rekap.remedial",
	// Settings
	"settings.read", "settings.update",
	// Audit
	"audit.read",
	// Analisis butir soal (BUG-01: sebelumnya hanya di migrasi 0008,
	// tidak di rbac/seed/kontrak sehingga proktor+guru 403)
	"analisis.read", "analisis.export",
	// Pengecualian darurat SoD (Fase 3) — admin saja.
	"sod.override",
	// UKK
	"ukk.read", "ukk.create", "ukk.update",
	// Event
	"event.read", "event.create", "event.update", "event.delete",
	// Backup
	"backup.create", "backup.restore",
	// Manajemen Pengguna staf (UC20) — admin saja. Tidak ada
	// permission pengguna.* untuk role lain: staf non-admin tidak
	// mengelola akun pengguna (SoD-6: pengelola user != pengelola
	// audit log; audit.read tetap dimiliki proktor).
	"pengguna.read", "pengguna.create", "pengguna.update", "pengguna.reset-password", "pengguna.toggle-aktif",
}

// rolePermissions adalah matriks role -> permission (selain admin).
// Admin selalu memiliki SEMUA permission (lihat HasPermission).
var rolePermissions = map[Role][]string{
	RoleProktor: {
		"master.jurusan.read", "master.kelas.read", "master.mapel.read", "master.guru.read", "master.siswa.read",
		"banksoal.read",
		"jadwal.read", "jadwal.create", "jadwal.update", "jadwal.delete",
		"jadwal.pause", "jadwal.resume", "jadwal.unblock", "jadwal.usulan-unblock",
		"ujian.monitor",
		"koreksi.read",
		"rekap.read", "rekap.export", "rekap.finalisasi", "rekap.remedial",
		"settings.read",
		"audit.read",
		"analisis.read", "analisis.export",
		"event.read", "event.create", "event.update", "event.delete",
		"backup.create", "backup.restore",
	},
	RoleGuru: {
		"master.jurusan.read", "master.kelas.read", "master.mapel.read", "master.guru.read", "master.siswa.read",
		"banksoal.read", "banksoal.create", "banksoal.update", "banksoal.delete", "banksoal.lock", "banksoal.import",
		"paket.read", "paket.create", "paket.update", "paket.delete",
		"jadwal.read", "jadwal.create", "jadwal.update", "jadwal.delete",
		"ujian.monitor",
		"koreksi.read", "koreksi.update",
		"rekap.read", "rekap.remedial",
		"settings.read",
		"analisis.read", "analisis.export",
	},
	RolePengawas: {
		"ujian.monitor",
		"jadwal.read",
		"jadwal.usulan-unblock",
		"rekap.read",
	},
	RoleSiswa: {
		"ujian.kerjakan",
		"rekap.read",
	},
	RolePengujiUKK: {
		"ukk.read", "ukk.create", "ukk.update",
		"rekap.read",
	},
}

// PermissionsForRole mengembalikan daftar permission untuk sebuah role.
// Admin mengembalikan AllPermissions.
func PermissionsForRole(role Role) []string {
	if role == RoleAdmin {
		out := make([]string, len(AllPermissions))
		copy(out, AllPermissions)
		return out
	}
	perms, ok := rolePermissions[role]
	if !ok {
		return nil
	}
	out := make([]string, len(perms))
	copy(out, perms)
	return out
}

// HasPermission melaporkan apakah role memiliki permission tertentu.
// Admin selalu true (deny-by-default untuk role lain yang tidak dikenal).
func HasPermission(role Role, perm string) bool {
	if role == RoleAdmin {
		return true
	}
	for _, p := range rolePermissions[role] {
		if p == perm {
			return true
		}
	}
	return false
}
