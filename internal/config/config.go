package config

import (
	"errors"
	"fmt"
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

func Load() (Config, error) {
	dir, err := projectsDir()
	if err != nil {
		return Config{}, err
	}

	return Config{
		Port:        getEnv("PORT", "8080"),
		LogLevel:    parseLogLevel(getEnv("LOG_LEVEL", "info")),
		ProjectsDir: dir,
		CorsOrigins: parseCorsOrigins(getEnv("CORS_ORIGINS", "*")),
	}, nil
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

func projectsDir() (string, error) {
	raw := os.Getenv("PERSISTENCE_DIR")
	if raw == "" {
		return "", errors.New("PERSISTENCE_DIR is not set: every project directory is created under it, so there is no sensible default to guess")
	}

	path, err := expandHome(raw, userHome())
	if err != nil {
		return "", fmt.Errorf("PERSISTENCE_DIR %q: %w", raw, err)
	}

	// A relative path resolves against the working directory, which in a
	// container is a writable layer that is discarded on the next start.
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("PERSISTENCE_DIR %q: must be an absolute path", raw)
	}

	if err := checkWritableDir(path); err != nil {
		return "", fmt.Errorf("PERSISTENCE_DIR %q: %w", path, err)
	}

	return path, nil
}

func userHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func expandHome(path, home string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path, nil
	}
	if home == "" {
		return "", errors.New("no home directory to expand ~ against")
	}
	return filepath.Join(home, strings.TrimPrefix(path, "~")), nil
}

// Writing a probe file rather than reading the mode bits, because the bits
// describe an owner the process may not be, and say nothing about a mount that
// is read-only.
func checkWritableDir(path string) error {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("directory does not exist")
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("not a directory")
	}

	probe, err := os.CreateTemp(path, ".writable-*")
	if err != nil {
		return fmt.Errorf("not writable: %w", err)
	}
	if err := probe.Close(); err != nil {
		return err
	}
	return os.Remove(probe.Name())
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
