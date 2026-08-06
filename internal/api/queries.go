package api

import (
	"log/slog"
	"net/http"
	"strconv"

	"nabu-storage/internal/domain"
)

const (
	defaultPageSize = 50
	maxPageSize     = 200
)

type Page[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func ProjectsHandler(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projects, err := domain.ListProjects(baseDir)
		if err != nil {
			slog.Error("failed to list projects", "baseDir", baseDir, "error", err)
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		page := positiveParam(r, "page", 1)
		size := min(positiveParam(r, "page_size", defaultPageSize), maxPageSize)

		writeJSON(w, http.StatusOK, Page[domain.Project]{
			Items:    slice(projects, page, size),
			Total:    len(projects),
			Page:     page,
			PageSize: size,
		})
	}
}

// A page beyond the end is an empty page rather than an error: a client holding a
// stale page number is asking a reasonable question about a corpus that shrank.
func slice[T any](items []T, page, size int) []T {
	start := (page - 1) * size
	if start >= len(items) {
		return []T{}
	}
	return items[start:min(start+size, len(items))]
}

func positiveParam(r *http.Request, name string, fallback int) int {
	value, err := strconv.Atoi(r.URL.Query().Get(name))
	if err != nil || value < 1 {
		return fallback
	}
	return value
}
