package main

import (
	"fmt"
	"log"
	"log/slog"
	net "net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"hermes-relay/internal/config"
	"hermes-relay/internal/handlers"
)

func main() {
	cfg := config.Load()
	setupLogger(cfg.LogLevel)

	slog.Info("Projects directory", "dir", cfg.ProjectsDir)

	r := chi.NewRouter()
	handlers.SetupRoutes(r, cfg.ProjectsDir, cfg.CorsOrigins)

	addr := fmt.Sprintf(":%s", cfg.Port)
	slog.Info("Server starting", "port", cfg.Port, "cors_origins", cfg.CorsOrigins)
	log.Fatal(net.ListenAndServe(addr, r))
}

func setupLogger(level slog.Level) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}
