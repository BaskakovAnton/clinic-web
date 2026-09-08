package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port          string
	DatabaseURL   string
	SessionSecret string
	SecureCookie  bool
}

func Load() (Config, error) {
	_ = loadDotEnv(".env")

	cfg := Config{
		Port:          getenv("PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		SessionSecret: os.Getenv("SESSION_SECRET"),
		SecureCookie:  getenv("SECURE_COOKIE", "") == "1" || getenv("SECURE_COOKIE", "") == "true" || getenv("APP_ENV", "") == "production",
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.SessionSecret == "" {
		return Config{}, fmt.Errorf("SESSION_SECRET is required")
	}
	if len(cfg.SessionSecret) < 16 {
		return Config{}, fmt.Errorf("SESSION_SECRET must be at least 16 characters")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		val = strings.Trim(val, `"'`)
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
	return sc.Err()
}
