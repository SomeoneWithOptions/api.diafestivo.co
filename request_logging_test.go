package main

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

func TestLoggingMiddlewareAsyncWithSynctest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var logs bytes.Buffer
		previousLogger := slog.Default()
		slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
		defer slog.SetDefault(previousLogger)

		cfg := config{logTimeout: time.Second}
		handler := loggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}), cfg)

		req := httptest.NewRequest(http.MethodGet, "/synctest", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		synctest.Wait()

		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
		gotLogs := logs.String()
		if !strings.Contains(gotLogs, "msg=request") || !strings.Contains(gotLogs, "path=/synctest") {
			t.Fatalf("log output missing request fields: %q", gotLogs)
		}
	})
}
