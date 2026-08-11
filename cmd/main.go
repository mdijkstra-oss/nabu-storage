package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"nabu-storage/internal/api"
	"nabu-storage/internal/config"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(healthcheck(os.Args[2:]))
	}
	os.Exit(serve())
}

func serve() int {
	cfg, err := config.Load()
	if err != nil {
		setupLogger(slog.LevelInfo)
		slog.Error("configuration rejected", "error", err)
		return 1
	}
	setupLogger(cfg.LogLevel)

	slog.Info("Projects directory", "dir", cfg.ProjectsDir)

	r := chi.NewRouter()
	api.SetupRoutes(r, cfg.ProjectsDir, cfg.CorsOrigins)

	addr := fmt.Sprintf(":%s", cfg.Port)
	slog.Info("Server starting", "port", cfg.Port, "cors_origins", cfg.CorsOrigins)

	server := &http.Server{Addr: addr, Handler: r, ReadHeaderTimeout: 10 * time.Second}
	return listen(server)
}

// Docker stops a container by signalling it and killing it ten seconds later, so
// the grace period has to end well inside that.
const shutdownGrace = 5 * time.Second

// Without this the Go runtime handles the signal itself and the process exits 2,
// which reads as a crash in `docker compose ps`. Shutdown leaves hijacked
// connections alone, so an open websocket does not hold the grace period open.
func listen(server *http.Server) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	failed := make(chan error, 1)
	go func() { failed <- server.ListenAndServe() }()

	select {
	case err := <-failed:
		slog.Error("server stopped", "error", err)
		return 1
	case <-ctx.Done():
	}

	slog.Info("Server stopping")
	grace, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()

	// A request still running at the deadline is dropped rather than kept: the
	// stop was asked for, and the alternative is being killed mid-write anyway.
	if err := server.Shutdown(grace); err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			slog.Error("shutdown failed", "error", err)
			return 1
		}
		slog.Warn("shutdown timed out; closing anyway", "after", shutdownGrace)
		_ = server.Close()
	}
	return 0
}

// Asks only whether this process is serving. The image is built from scratch and
// holds one file, so a container HEALTHCHECK has no shell, no curl and no wget to
// call instead.
func healthcheck(args []string) int {
	fs := flag.NewFlagSet("healthcheck", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:8080", "Address to check.")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	client := &http.Client{Timeout: 3 * time.Second}
	response, err := client.Get("http://" + *addr + "/health")
	if err != nil {
		fmt.Fprintf(os.Stderr, "reach %s: %v\n", *addr, err)
		return 1
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "%s answered %s\n", *addr, response.Status)
		return 1
	}
	return 0
}

func setupLogger(level slog.Level) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}
