package auth

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"
)

// hashDenganParam membuat hash PHC Argon2id dengan parameter eksplisit
// (hanya untuk tes — produksi memakai HashPassword).
func hashDenganParam(password string, salt []byte, t, m uint32, p uint8) string {
	hash := argon2.IDKey([]byte(password), salt, t, m, p, 32)
	b64 := base64.RawStdEncoding.EncodeToString
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		m, t, p, b64(salt), b64(hash))
}

func TestHashVerifyRoundtrip(t *testing.T) {
	hash, err := HashPassword("PasswordKuat123")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$m=65536,t=1,p=4$") {
		t.Fatalf("format hash salah: %s", hash)
	}
	ok, err := VerifyPassword("PasswordKuat123", hash)
	if err != nil || !ok {
		t.Fatalf("verify benar = %v, err = %v", ok, err)
	}
	ok, err = VerifyPassword("password-salah", hash)
	if err != nil || ok {
		t.Fatalf("verify salah = %v, err = %v", ok, err)
	}
	// Dua hash password sama harus berbeda (salt acak).
	hash2, _ := HashPassword("PasswordKuat123")
	if hash == hash2 {
		t.Fatal("salt harus acak: dua hash identik")
	}
}

func TestVerifyPasswordFormatSalah(t *testing.T) {
	if _, err := VerifyPassword("x", "bukan-hash"); err == nil {
		t.Fatal("hash format salah harus error")
	}
	if _, err := VerifyPassword("x", "$argon2id$v=19$m=1,t=1,p=1$!!!$???"); err == nil {
		t.Fatal("hash base64 tidak valid harus error")
	}
}

// TestVerifyPasswordParameterBerbeda memastikan VerifyPassword mem-parse
// parameter Argon2id dari string hash PHC (m/t/p), BUKAN hardcode.
// Hash siswa hasil import memakai m=65536,t=3,p=2, sedangkan hash admin
// (HashPassword) memakai t=1,p=4 — keduanya harus bisa login.
func TestVerifyPasswordParameterBerbeda(t *testing.T) {
	salt := []byte("1234567890123456") // 16 byte, deterministik untuk tes
	hashSiswa := hashDenganParam("1234567890", salt, 3, 65536, 2)
	if !strings.HasPrefix(hashSiswa, "$argon2id$v=19$m=65536,t=3,p=2$") {
		t.Fatalf("format hash siswa salah: %s", hashSiswa)
	}
	if ok, err := VerifyPassword("1234567890", hashSiswa); err != nil || !ok {
		t.Fatalf("hash siswa t=3,p=2 harus terverifikasi: ok=%v err=%v", ok, err)
	}
	if ok, _ := VerifyPassword("password-salah", hashSiswa); ok {
		t.Fatal("password salah harus ditolak untuk hash siswa")
	}
	hashAdmin := hashDenganParam("AdminKuat123", salt, 1, 65536, 4)
	if ok, err := VerifyPassword("AdminKuat123", hashAdmin); err != nil || !ok {
		t.Fatalf("hash admin t=1,p=4 harus terverifikasi: ok=%v err=%v", ok, err)
	}
}
