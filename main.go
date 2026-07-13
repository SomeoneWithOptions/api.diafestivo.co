package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const defaultPort = "3002"

type config struct {
	port        string
	ipInfoToken string
	myCIDR      string

	readHeaderTimeout time.Duration
	readTimeout       time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
	shutdownTimeout   time.Duration
	logTimeout        time.Duration
}

func main() {
	setupLogger()

	cfg, err := loadConfig()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	server := newHTTPServer(cfg)
	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("starting http server", "addr", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-shutdownSignals:
		slog.Info("shutting down http server", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), cfg.shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			slog.Error("http server graceful shutdown failed", "error", err)
			if closeErr := server.Close(); closeErr != nil {
				slog.Error("http server forced close failed", "error", closeErr)
			}
		}
		if err := <-serverErrors; err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server stopped", "error", err)
			os.Exit(1)
		}
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server stopped", "error", err)
			os.Exit(1)
		}
	}
}

func setupLogger() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
}

func loadConfig() (config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	if err := validatePort(port); err != nil {
		return config{}, err
	}

	return config{
		port:              port,
		ipInfoToken:       os.Getenv("IP_INFO_TOKEN"),
		myCIDR:            os.Getenv("MY_CIDR"),
		readHeaderTimeout: 5 * time.Second,
		readTimeout:       15 * time.Second,
		writeTimeout:      30 * time.Second,
		idleTimeout:       60 * time.Second,
		shutdownTimeout:   10 * time.Second,
		logTimeout:        3 * time.Second,
	}, nil
}

func validatePort(port string) error {
	parsedPort, err := strconv.Atoi(port)
	if err != nil || parsedPort < 1 || parsedPort > 65535 {
		return fmt.Errorf("invalid PORT %q: must be an integer from 1 to 65535", port)
	}
	return nil
}

func newHTTPServer(cfg config) *http.Server {
	return &http.Server{
		Addr:              net.JoinHostPort("", cfg.port),
		Handler:           loggingMiddleware(newServeMux(), cfg),
		ReadHeaderTimeout: cfg.readHeaderTimeout,
		ReadTimeout:       cfg.readTimeout,
		WriteTimeout:      cfg.writeTimeout,
		IdleTimeout:       cfg.idleTimeout,
	}
}
