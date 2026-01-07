package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Port           string
	LogLevel       slog.Level
	PersistenceDir string
	CorsOrigins    []string
}

func Load() Config {
	return Config{
		Port:           getEnv("PORT", "8080"),
		LogLevel:       parseLogLevel(getEnv("LOG_LEVEL", "info")),
		PersistenceDir: getPersistenceDir(),
		CorsOrigins:    parseCorsOrigins(getEnv("CORS_ORIGINS", "*")),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getPersistenceDir() string {
	if dir := os.Getenv("PERSISTENCE_DIR"); dir != "" {
		return dir
	}

	home, err := os.UserHomeDir()
	if err != nil {
		slog.Warn("failed to get user home directory, using relative path", "error", err)
		return "persistence-data"
	}
	return filepath.Join(home, "Documents", "nabu-persistence")
}

func parseCorsOrigins(origins string) []string {
	if origins == "" {
		return []string{}
	}
	parts := strings.Split(origins, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
