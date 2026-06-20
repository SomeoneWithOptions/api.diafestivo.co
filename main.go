package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

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
	handlerOptions := &slog.HandlerOptions{Level: slog.LevelInfo}
	stdoutHandler := slog.NewJSONHandler(os.Stdout, handlerOptions)
	stderrHandler := slog.NewJSONHandler(os.Stderr, handlerOptions)
	slog.SetDefault(slog.New(newLevelSplitHandler(stdoutHandler, stderrHandler, slog.LevelError)))
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
