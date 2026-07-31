package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"nabu-storage/internal/api/websocket"
	"nabu-storage/internal/lib/utils"
)

func SetupRoutes(r chi.Router, baseDir string, corsOrigins []string) {
	r.Use(slogLogger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.StripSlashes)
	r.Use(buildCorsHandler(corsOrigins))

	r.With(DecompressGzip).Post("/commands/{projectId}", CommandHandler(baseDir))
	r.Get("/ws/{projectId}", websocket.Handler(baseDir))

	utils.MustNotError(chi.Walk(r, logRoute))
}

func logRoute(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
	slog.Debug("registered route", "method", method, "route", route)
	return nil
}

func buildCorsHandler(allowedOrigins []string) func(http.Handler) http.Handler {
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}

	return cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})
}

func slogLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		defer func() {
			slog.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration_ms", time.Since(start).Milliseconds(),
			)
		}()

		next.ServeHTTP(ww, r)
	})
}
