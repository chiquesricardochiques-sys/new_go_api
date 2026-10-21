package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port        string
	DatabaseDSN string
	AuthBaseURL string
}

func Load() (Config, error) {
	loadEnv(".env")

	cfg := Config{
		Port:        getEnv("API_PORT", "8080"),
		DatabaseDSN: os.Getenv("DATABASE_DSN"),
		AuthBaseURL: strings.TrimRight(os.Getenv("AUTH_BASE_URL"), "/"),
	}

	if cfg.DatabaseDSN == "" {
		return Config{}, fmt.Errorf("DATABASE_DSN não configurado")
	}

	return cfg, nil
}

func loadEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)

		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
