package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/SomeoneWithOptions/api.diafestivo.co/holiday"
)

type requestLogData struct {
	method     string
	url        string
	path       string
	proto      string
	requestIP  string
	status     int
	duration   time.Duration
	loggedTime time.Time
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(body)
}

var invalidCIDRLogOnce sync.Once

func loggingMiddleware(next http.Handler, cfg config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)

		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}

		data := requestLogData{
			method:     r.Method,
			url:        r.URL.String(),
			path:       r.URL.Path,
			proto:      r.Header.Get("X-Forwarded-Proto"),
			requestIP:  requestIP(r),
			status:     status,
			duration:   time.Since(started),
			loggedTime: holiday.NowInCOT(),
		}

		ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), cfg.logTimeout)
		go func() {
			defer cancel()
			logRequest(ctx, data, cfg)
		}()
	})
}

func requestIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		return strings.TrimSpace(strings.Split(forwardedFor, ",")[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func logRequest(ctx context.Context, data requestLogData, cfg config) {
	requestAddr, err := netip.ParseAddr(data.requestIP)
	if err != nil {
		logLocalRequest(data, "invalid_or_missing_ip")
		return
	}

	if cfg.myCIDR != "" {
		isOwnIP, err := IP(data.requestIP).IsInCIDR(cfg.myCIDR)
		if err != nil {
			invalidCIDRLogOnce.Do(func() {
				slog.Error("invalid MY_CIDR", "cidr", cfg.myCIDR, "error", err)
			})
		} else if isOwnIP {
			return
		}
	}

	if cfg.ipInfoToken == "" {
		logLocalRequest(data, "missing_ip_info_token")
		return
	}

	ipInfo, err := IP(requestAddr.String()).FetchIPInfoContext(ctx)
	if err != nil {
		if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			slog.Error("failed to fetch ip info", "error", err, "request_ip", data.requestIP)
		}
		logLocalRequest(data, "ip_info_unavailable")
		return
	}

	slog.Info("request",
		"time", data.loggedTime.Format("02-01-2006:15:04:05"),
		"method", data.method,
		"url", data.url,
		"path", data.path,
		"proto", data.proto,
		"status", data.status,
		"duration_ms", data.duration.Milliseconds(),
		"ip", ipInfo.IP,
		"city", ipInfo.City,
		"region", ipInfo.Region,
		"country", ipInfo.Country,
	)
}

func logLocalRequest(data requestLogData, reason string) {
	slog.Info("request",
		"time", data.loggedTime.Format("02-01-2006:15:04:05"),
		"method", data.method,
		"url", data.url,
		"path", data.path,
		"proto", data.proto,
		"status", data.status,
		"duration_ms", data.duration.Milliseconds(),
		"ip", data.requestIP,
		"geo", reason,
	)
}
