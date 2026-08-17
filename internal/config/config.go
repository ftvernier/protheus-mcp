package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ProtheusProcess          string
	DBHost                   string
	DBPort                   int
	DBName                   string
	DBUser                   string
	DBPassword               string
	DBEncrypt                string
	DBTrustServerCertificate string
	QueryTimeout             time.Duration
}

func Load() Config {
	return Config{
		ProtheusProcess:          env("PROTHEUS_PROCESS", "appserver"),
		DBHost:                   strings.TrimSpace(os.Getenv("DB_HOST")),
		DBPort:                   envInt("DB_PORT", 1433),
		DBName:                   strings.TrimSpace(os.Getenv("DB_NAME")),
		DBUser:                   strings.TrimSpace(os.Getenv("DB_USER")),
		DBPassword:               os.Getenv("DB_PASSWORD"),
		DBEncrypt:                env("DB_ENCRYPT", "true"),
		DBTrustServerCertificate: env("DB_TRUST_SERVER_CERTIFICATE", "false"),
		QueryTimeout:             time.Duration(envInt("QUERY_TIMEOUT_SECONDS", 5)) * time.Second,
	}
}

func (c Config) DatabaseConfigured() bool {
	return c.DBHost != "" && c.DBName != "" && c.DBUser != ""
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
