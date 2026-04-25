package handlers

import (
	"log/slog"
	net "net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"hermes-relay/internal/handlers/http"
	"hermes-relay/internal/handlers/http/websocket"
	"hermes-relay/internal/lib/utils"
)

func SetupRoutes(r chi.Router, baseDir string, hub *websocket.Hub, corsOrigins []string) {
	r.Use(slogLogger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.StripSlashes)
	r.Use(buildCorsHandler(corsOrigins))

	r.With(http.DecompressGzip).Post("/commands/{projectId}", http.CommandHandler(baseDir))
	r.Get("/ws/{projectId}", websocket.Handler(hub, baseDir))

	utils.MustNotError(chi.Walk(r, logRoute))
}

func logRoute(method, route string, _ net.Handler, _ ...func(net.Handler) net.Handler) error {
	slog.Debug("registered route", "method", method, "route", route)
	return nil
}

func buildCorsHandler(allowedOrigins []string) func(net.Handler) net.Handler {
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}

	return cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})
}

func slogLogger(next net.Handler) net.Handler {
	return net.HandlerFunc(func(w net.ResponseWriter, r *net.Request) {
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
