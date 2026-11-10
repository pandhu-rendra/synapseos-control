package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/synapseos/synapseos-control/apps/api/internal/docker"
)

type config struct {
	address        string
	dockerHost     string
	requestTimeout time.Duration
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config{
		address:        getenv("SYNAPSEOS_API_ADDRESS", ":8080"),
		dockerHost:     getenv("DOCKER_HOST", "unix:///var/run/docker.sock"),
		requestTimeout: 10 * time.Second,
	}
	dockerProvider, err := docker.NewManager(docker.Settings{
		Host:    cfg.dockerHost,
		Timeout: int(cfg.requestTimeout / time.Second),
	})
	if err != nil {
		logger.Error("docker provider initialization failed", "error", err)
		os.Exit(1)
	}
	defer dockerProvider.Close()

	server := &http.Server{
		Addr:              cfg.address,
		Handler:           newRouter(dockerProvider),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("api server listening", "address", cfg.address)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api server stopped unexpectedly", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Error("api server shutdown failed", "error", err)
		os.Exit(1)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
