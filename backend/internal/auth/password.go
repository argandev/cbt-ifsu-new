// Password hashing Argon2id (Tim 1).
//
// Parameter sesuai kontrak: time=1, memory=64 MiB (65536 KiB), threads=4,
// keyLen=32, salt acak 16 byte. Format penyimpanan:
// $argon2id$v=19$m=65536,t=1,p=4$<base64-salt-tanpa-padding>$<base64-hash-tanpa-padding>
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 1
	argonMemory  = 65536 // KiB = 64 MiB
	argonThreads = 4
	argonKeyLen  = 32
	argonSaltLen = 16
)

var errInvalidHash = errors.New("format hash password tidak valid")

// HashPassword meng-hash password plaintext menjadi string Argon2id.
func HashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("gagal membangkitkan salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	b64 := base64.RawStdEncoding.EncodeToString
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory, argonTime, argonThreads, b64(salt), b64(hash)), nil
}

// VerifyPassword memeriksa password terhadap hash Argon2id tersimpan.
// Mengembalikan false (tanpa error) bila tidak cocok; error bila format
// hash tidak dapat diparse.
func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	// ["", "argon2id", "v=19", "m=...,t=...,p=...", salt, hash]
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, errInvalidHash
	}
	var memory, timeCost uint32
	var threads uint8
	for _, kv := range strings.Split(parts[3], ",") {
		k, v, found := strings.Cut(kv, "=")
		if !found {
			return false, errInvalidHash
		}
		n, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return false, errInvalidHash
		}
		switch k {
		case "m":
			memory = uint32(n)
		case "t":
			timeCost = uint32(n)
		case "p":
			threads = uint8(n)
		default:
			return false, errInvalidHash
		}
	}
	b64 := base64.RawStdEncoding.DecodeString
	salt, err := b64(parts[4])
	if err != nil {
		return false, errInvalidHash
	}
	expected, err := b64(parts[5])
	if err != nil {
		return false, errInvalidHash
	}
	computed := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, uint32(len(expected)))
	return subtle.ConstantTimeCompare(computed, expected) == 1, nil
}
