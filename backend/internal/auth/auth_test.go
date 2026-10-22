package auth

import (
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Known-answer test for PBKDF2-HMAC-SHA256 (P="password", S="salt", c=1).
func TestPBKDF2KnownVector(t *testing.T) {
	got := hex.EncodeToString(pbkdf2SHA256([]byte("password"), []byte("salt"), 1, 32))
	want := "120fb6cffcf8b32c43e7225256c4f837a86548c92ccc35480805987cb70be17b"
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

// A fixed hash in the stored format: checks verification against a known value.
const knownHash = "$pbkdf2-sha256$29000$AAECAwQFBgcICQoLDA0ODw$1dv1PW93ILN8auHyNNA7Gz1XqZJFK39ThafYVcEfKtQ"

func TestVerifiesKnownHash(t *testing.T) {
	if !VerifyPassword("Secret@123", knownHash) {
		t.Fatal("expected known hash to verify")
	}
	if VerifyPassword("secret@123", knownHash) {
		t.Fatal("wrong password must not verify")
	}
}

func TestHashRoundTrip(t *testing.T) {
	h, err := HashPassword("hunter2!")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(h, "$pbkdf2-sha256$29000$") {
		t.Fatalf("unexpected format %s", h)
	}
	if !VerifyPassword("hunter2!", h) || VerifyPassword("hunter3!", h) {
		t.Fatal("round trip failed")
	}
	h2, _ := HashPassword("hunter2!")
	if h == h2 {
		t.Fatal("salts must be random")
	}
}

func TestMalformedHashes(t *testing.T) {
	for _, h := range []string{"", "plain", "$bcrypt$x$y$z", "$pbkdf2-sha256$abc$AA$AA", "$pbkdf2-sha256$1$!!$AA"} {
		if VerifyPassword("x", h) {
			t.Errorf("malformed hash %q verified", h)
		}
	}
}

func TestJWT(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	m := NewTokenManager("a-long-test-secret-key", time.Hour)
	m.Now = func() time.Time { return now }

	tok, err := m.Issue("priya@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if email, err := m.Parse(tok); err != nil || email != "priya@example.com" {
		t.Fatalf("Parse = %q, %v", email, err)
	}

	// Expired
	m.Now = func() time.Time { return now.Add(2 * time.Hour) }
	if _, err := m.Parse(tok); err == nil {
		t.Fatal("expired token accepted")
	}
	m.Now = func() time.Time { return now }

	// Wrong secret
	other := NewTokenManager("a-different-secret-key", time.Hour)
	other.Now = m.Now
	if _, err := other.Parse(tok); err == nil {
		t.Fatal("token signed with another secret accepted")
	}

	// "alg: none" must be rejected
	unsigned, _ := jwt.NewWithClaims(jwt.SigningMethodNone,
		jwt.MapClaims{"sub": "x@y.com", "exp": now.Add(time.Hour).Unix()}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if _, err := m.Parse(unsigned); err == nil {
		t.Fatal("unsigned token accepted")
	}

	// Missing exp must be rejected
	noExp, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "x@y.com"}).SignedString(m.Secret)
	if _, err := m.Parse(noExp); err == nil {
		t.Fatal("token without exp accepted")
	}
}
