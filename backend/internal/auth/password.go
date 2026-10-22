// Package auth handles password hashing and JWT tokens.
//
// Passwords use salted PBKDF2-SHA256 (29,000 rounds), stored in the standard
// "$pbkdf2-sha256$<rounds>$<salt>$<hash>" format.
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

const (
	pbkdf2Rounds = 29000 // work factor: slows down brute-force attacks
	saltLen      = 16
	keyLen       = 32
)

// pbkdf2SHA256 implements PBKDF2 (RFC 8018) with HMAC-SHA256.
func pbkdf2SHA256(password, salt []byte, iter, keyLen int) []byte {
	prf := hmac.New(sha256.New, password)
	hLen := prf.Size()
	blocks := (keyLen + hLen - 1) / hLen
	out := make([]byte, 0, blocks*hLen)
	buf := make([]byte, 4)
	u := make([]byte, hLen)
	for block := 1; block <= blocks; block++ {
		prf.Reset()
		prf.Write(salt)
		binary.BigEndian.PutUint32(buf, uint32(block))
		prf.Write(buf)
		u = prf.Sum(u[:0])
		t := make([]byte, hLen)
		copy(t, u)
		for i := 1; i < iter; i++ {
			prf.Reset()
			prf.Write(u)
			u = prf.Sum(u[:0])
			for j := range t {
				t[j] ^= u[j]
			}
		}
		out = append(out, t...)
	}
	return out[:keyLen]
}

// "adapted base64": standard alphabet, '.' instead of '+', no padding.
func ab64Encode(b []byte) string {
	return strings.ReplaceAll(base64.RawStdEncoding.EncodeToString(b), "+", ".")
}

func ab64Decode(s string) ([]byte, error) {
	return base64.RawStdEncoding.DecodeString(strings.ReplaceAll(s, ".", "+"))
}

// HashPassword returns a salted pbkdf2_sha256 hash.
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return encodeHash(password, salt, pbkdf2Rounds), nil
}

func encodeHash(password string, salt []byte, rounds int) string {
	dk := pbkdf2SHA256([]byte(password), salt, rounds, keyLen)
	return fmt.Sprintf("$pbkdf2-sha256$%d$%s$%s", rounds, ab64Encode(salt), ab64Encode(dk))
}

// VerifyPassword checks a password against a stored hash in constant time.
func VerifyPassword(password, stored string) bool {
	parts := strings.Split(stored, "$")
	// "", "pbkdf2-sha256", rounds, salt, hash
	if len(parts) != 5 || parts[1] != "pbkdf2-sha256" {
		return false
	}
	rounds, err := strconv.Atoi(parts[2])
	if err != nil || rounds < 1 {
		return false
	}
	salt, err := ab64Decode(parts[3])
	if err != nil {
		return false
	}
	want, err := ab64Decode(parts[4])
	if err != nil || len(want) == 0 {
		return false
	}
	got := pbkdf2SHA256([]byte(password), salt, rounds, len(want))
	return subtle.ConstantTimeCompare(got, want) == 1
}
