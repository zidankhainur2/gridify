package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Port               string
	Env                string
	DatabaseURL        string
	SessionSecret      string
	CORSAllowedOrigins []string
	SMTPURL            string
}

func Load() (*Config, error) {
	loadDotEnv()

	port := getEnv("PORT", "8080")
	env := getEnv("ENV", "development")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:admin@localhost:5432/gridify?sslmode=disable")
	sessionSecret := getEnv("SESSION_SECRET", "gridify_secure_session_secret_key_minimum_32_bytes_test")
	corsRaw := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000")
	smtpURL := getEnv("SMTP_URL", "smtp://localhost:1025")

	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if len(sessionSecret) < 32 {
		return nil, fmt.Errorf("SESSION_SECRET must be at least 32 characters long")
	}

	var corsOrigins []string
	for _, origin := range strings.Split(corsRaw, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			corsOrigins = append(corsOrigins, trimmed)
		}
	}

	return &Config{
		Port:               port,
		Env:                env,
		DatabaseURL:        dbURL,
		SessionSecret:      sessionSecret,
		CORSAllowedOrigins: corsOrigins,
		SMTPURL:            smtpURL,
	}, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// loadDotEnv tries to find and parse .env files in the current dir or parent dirs.
func loadDotEnv() {
	candidates := []string{
		".env",
		"../.env",
		"../../.env",
	}

	for _, relPath := range candidates {
		absPath, err := filepath.Abs(relPath)
		if err != nil {
			continue
		}
		if file, err := os.Open(absPath); err == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					k := strings.TrimSpace(parts[0])
					v := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
					if os.Getenv(k) == "" {
						os.Setenv(k, v)
					}
				}
			}
			return
		}
	}
}
