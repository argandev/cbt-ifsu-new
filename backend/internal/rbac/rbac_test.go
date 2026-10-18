package rbac

import "testing"

// TestMatriksSesuaiKontrak memastikan matriks di kode == matriks di
// docs/KONTRAK.md == seed Tim 6 (backend/db/seed/seed.sql).
// Bila salah satu berubah, ketiganya harus diubah bersamaan.
func TestMatriksSesuaiKontrak(t *testing.T) {
	// UC20: 67 = 62 + pengguna.read + pengguna.create + pengguna.update
	// + pengguna.reset-password + pengguna.toggle-aktif (admin saja).
	if got := len(AllPermissions); got != 67 {
		t.Fatalf("jumlah permission = %d, mau 67", got)
	}
	cases := map[Role]int{
		RoleAdmin:      67,
		RoleProktor:    30,
		RoleGuru:       27,
		RolePengawas:   4,
		RoleSiswa:      2,
		RolePengujiUKK: 4,
	}
	for role, want := range cases {
		if got := len(PermissionsForRole(role)); got != want {
			t.Errorf("role %s: %d permission, mau %d", role, got, want)
		}
	}
	// Admin lolos semua.
	for _, p := range AllPermissions {
		if !HasPermission(RoleAdmin, p) {
			t.Errorf("admin seharusnya punya %s", p)
		}
	}
	// Batasan penting (ANALISIS Bab 11): proktor & pengawas tidak boleh
	// melihat isi soal (tidak ada banksoal.* kecuali banksoal.read
	// metadata untuk proktor); pengawas tidak boleh melihat token
	// (tidak ada permission token apa pun untuk pengawas).
	if HasPermission(RolePengawas, "banksoal.read") {
		t.Error("pengawas tidak boleh punya banksoal.read")
	}
	if HasPermission(RoleProktor, "banksoal.create") {
		t.Error("proktor tidak boleh membuat bank soal")
	}
	if HasPermission(RoleProktor, "paket.read") {
		t.Error("proktor tidak boleh melihat paket (tidak boleh melihat isi soal)")
	}
	if !HasPermission(RolePengawas, "jadwal.usulan-unblock") {
		t.Error("pengawas harus boleh mengusulkan unblock")
	}
	if !HasPermission(RoleProktor, "rekap.finalisasi") {
		t.Error("proktor harus boleh finalisasi rekap")
	}
	if !HasPermission(RoleGuru, "rekap.remedial") {
		t.Error("guru harus boleh remedial")
	}
	if HasPermission(RoleSiswa, "ujian.monitor") {
		t.Error("siswa tidak boleh monitor")
	}
	if HasPermission(RoleGuru, "settings.update") {
		t.Error("guru tidak boleh ubah settings")
	}
	// Fase 3: analisis untuk proktor+guru (BUG-01); sod.override admin saja.
	if !HasPermission(RoleProktor, "analisis.read") || !HasPermission(RoleProktor, "analisis.export") {
		t.Error("proktor harus punya analisis.read + analisis.export")
	}
	if !HasPermission(RoleGuru, "analisis.read") || !HasPermission(RoleGuru, "analisis.export") {
		t.Error("guru harus punya analisis.read + analisis.export")
	}
	if HasPermission(RoleProktor, "sod.override") || HasPermission(RoleGuru, "sod.override") {
		t.Error("sod.override hanya untuk admin")
	}
	// UC20: pengguna.* hanya untuk admin.
	for _, p := range []string{"pengguna.read", "pengguna.create", "pengguna.update", "pengguna.reset-password", "pengguna.toggle-aktif"} {
		if !HasPermission(RoleAdmin, p) {
			t.Errorf("admin harus punya %s", p)
		}
		for _, r := range []Role{RoleProktor, RoleGuru, RolePengawas, RoleSiswa, RolePengujiUKK} {
			if HasPermission(r, p) {
				t.Errorf("role %s tidak boleh punya %s", r, p)
			}
		}
	}
	if HasPermission(Role("role_tidak_dikenal"), "rekap.read") {
		t.Error("role tak dikenal harus ditolak (deny-by-default)")
	}
}
