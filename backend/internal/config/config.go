// Package config reads settings from environment variables (and an
// optional .env file).
package config

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port           string
	DatabaseURL    string // empty = use the in-memory store (demo mode)
	JWTSecret      string
	TokenTTL       time.Duration
	AllowedOrigins []string
	AdminEmail     string // optional: create/promote this admin at startup
	AdminPassword  string
}

var defaultOrigins = "http://localhost:5173,https://ai-payment-support-system.vercel.app"

// Load reads .env (if present) and then the environment.
func Load() (Config, error) {
	loadDotEnv(".env")
	ttlMin, err := strconv.Atoi(getenv("ACCESS_TOKEN_EXPIRE_MINUTES", "60"))
	if err != nil || ttlMin <= 0 {
		return Config{}, errors.New("ACCESS_TOKEN_EXPIRE_MINUTES must be a positive number")
	}
	c := Config{
		Port:           getenv("PORT", "8080"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		JWTSecret:      os.Getenv("JWT_SECRET_KEY"),
		TokenTTL:       time.Duration(ttlMin) * time.Minute,
		AllowedOrigins: strings.Split(getenv("CORS_ORIGINS", defaultOrigins), ","),
		AdminEmail:     os.Getenv("ADMIN_EMAIL"),
		AdminPassword:  os.Getenv("ADMIN_PASSWORD"),
	}
	if alg := getenv("JWT_ALGORITHM", "HS256"); alg != "HS256" {
		return Config{}, errors.New("only JWT_ALGORITHM=HS256 is supported")
	}
	if len(c.JWTSecret) < 16 {
		return Config{}, errors.New("JWT_SECRET_KEY must be set (at least 16 characters)")
	}
	return c, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// loadDotEnv sets KEY=VALUE lines from a file without overriding
// variables that are already set in the environment.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k, v = strings.TrimSpace(k), strings.Trim(strings.TrimSpace(v), `"'`)
		if _, exists := os.LookupEnv(k); !exists {
			os.Setenv(k, v)
		}
	}
}
