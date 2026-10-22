package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenManager issues and checks HS256 JWTs with claims {"sub": email, "exp": ...}.
type TokenManager struct {
	Secret []byte
	TTL    time.Duration
	Now    func() time.Time
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{Secret: []byte(secret), TTL: ttl, Now: time.Now}
}

var ErrInvalidToken = errors.New("invalid or expired token")

// Issue creates a signed token for the given email.
func (m *TokenManager) Issue(email string) (string, error) {
	claims := jwt.MapClaims{
		"sub": email,
		"exp": m.Now().Add(m.TTL).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.Secret)
}

// Parse validates the signature, algorithm and expiry, and returns the email.
func (m *TokenManager) Parse(token string) (string, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims,
		func(t *jwt.Token) (any, error) { return m.Secret, nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(m.Now),
	)
	if err != nil {
		return "", ErrInvalidToken
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", ErrInvalidToken
	}
	return sub, nil
}
