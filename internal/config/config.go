package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Port        string
	LogLevel    slog.Level
	ProjectsDir string
	CorsOrigins []string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		LogLevel:    parseLogLevel(getEnv("LOG_LEVEL", "info")),
		ProjectsDir: getProjectsDir(),
		CorsOrigins: parseCorsOrigins(getEnv("CORS_ORIGINS", "*")),
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

const (
	defaultProjectsDir  = "~/Documents/nabu-persistence"
	fallbackProjectsDir = "projects"
)

func getProjectsDir() string {
	return expandHome(getEnv("PERSISTENCE_DIR", defaultProjectsDir), userHome())
}

func userHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Warn("failed to get user home directory", "error", err)
		return ""
	}
	return home
}

func expandHome(path, home string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path
	}
	if home == "" {
		return fallbackProjectsDir
	}
	return filepath.Join(home, strings.TrimPrefix(path, "~"))
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
