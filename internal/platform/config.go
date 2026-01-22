package platform

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName     string
	HTTPPort        int
	ShutdownTimeout time.Duration

	// Stage 0 readiness is a TCP dial to PostgreSQL.
	PostgresAddr  string // host:port, e.g. postgres:5432
	CheckPostgres bool

	LogLevel string
}

// LoadConfig reads configuration from environment variables.
func LoadConfig(defaultService string) Config {
	return Config{
		ServiceName:     getenv("SERVICE_NAME", defaultService),
		HTTPPort:        getenvInt("HTTP_PORT", 8080),
		ShutdownTimeout: getenvDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		PostgresAddr:    getenv("POSTGRES_ADDR", "postgres:5432"),
		CheckPostgres:   getenvBool("CHECK_POSTGRES", true),
		LogLevel:        getenv("LOG_LEVEL", "INFO"),
	}
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getenvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func getenvBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func getenvDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
