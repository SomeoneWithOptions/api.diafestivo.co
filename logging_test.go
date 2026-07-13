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

func TestIsInCIDR(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		cidr    string
		want    bool
		wantErr bool
	}{
		{name: "empty cidr", ip: "203.0.113.10", cidr: "", want: false},
		{name: "ipv4 in cidr", ip: "203.0.113.10", cidr: "203.0.113.0/24", want: true},
		{name: "ipv4 out of cidr", ip: "203.0.114.10", cidr: "203.0.113.0/24", want: false},
		{name: "ipv6 in cidr", ip: "2001:db8::1", cidr: "2001:db8::/32", want: true},
		{name: "invalid ip", ip: "bad", cidr: "203.0.113.0/24", wantErr: true},
		{name: "invalid cidr", ip: "203.0.113.10", cidr: "bad", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isInCIDR(tt.ip, tt.cidr)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("isInCIDR = %v, want %v", got, tt.want)
			}
		})
	}
}

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
