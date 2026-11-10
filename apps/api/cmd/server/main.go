package main

import (
	"bufio"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/synapseos/synapseos-control/apps/api/internal/docker"
)

type config struct {
	address        string
	apiToken       string
	dockerHost     string
	requestTimeout time.Duration
}

func main() {
	loadDotEnv(".env")
	loadDotEnv(filepath.Join("apps", "api", ".env"))
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config{
		address:        getenv("SYNAPSEOS_API_ADDRESS", "127.0.0.1:8080"),
		apiToken:       os.Getenv("SYNAPSEOS_API_TOKEN"),
		dockerHost:     getenv("DOCKER_HOST", "unix:///var/run/docker.sock"),
		requestTimeout: 10 * time.Second,
	}
	if cfg.apiToken == "" {
		logger.Error("SYNAPSEOS_API_TOKEN must be configured")
		os.Exit(1)
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
		Handler:           withAPIToken(newRouter(dockerProvider), cfg.apiToken),
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

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found || strings.TrimSpace(key) == "" {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}
