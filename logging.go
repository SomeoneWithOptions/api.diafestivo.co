package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const maxIPInfoResponseBytes = 1 << 20

var (
	errMissingIPInfoToken = errors.New("missing IP_INFO_TOKEN")
	errInvalidIP          = errors.New("invalid ip address")
	ipInfoHTTPClient      = &http.Client{Timeout: 3 * time.Second}
	invalidCIDRLogOnce    sync.Once
)

type ipInfo struct {
	IP      string `json:"ip,omitempty"`
	City    string `json:"city,omitempty"`
	Region  string `json:"region,omitempty"`
	Country string `json:"country,omitempty"`
}

type requestLogData struct {
	method    string
	url       string
	path      string
	proto     string
	requestIP string
	status    int
	duration  time.Duration
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
			method: r.Method, url: r.URL.String(), path: r.URL.Path,
			proto: r.Header.Get("X-Forwarded-Proto"), requestIP: requestIP(r),
			status: status, duration: time.Since(started),
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
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func logRequest(ctx context.Context, data requestLogData, cfg config) {
	if data.status < http.StatusOK || data.status >= http.StatusMultipleChoices {
		logLocalRequest(data, "non_success_status")
		return
	}

	requestAddr, err := netip.ParseAddr(data.requestIP)
	if err != nil {
		logLocalRequest(data, "invalid_or_missing_ip")
		return
	}

	if cfg.myCIDR != "" {
		isOwnIP, err := isInCIDR(data.requestIP, cfg.myCIDR)
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

	info, err := fetchIPInfoContext(ctx, requestAddr.String())
	if err != nil {
		if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			slog.Error("failed to fetch ip info", "error", err, "request_ip", data.requestIP)
		}
		logLocalRequest(data, "ip_info_unavailable")
		return
	}

	slog.Info("request",
		"method", data.method, "url", data.url, "path", data.path, "proto", data.proto,
		"status", data.status, "duration_ms", data.duration.Milliseconds(),
		"ip", info.IP, "city", info.City, "region", info.Region, "country", info.Country,
	)
}

func logLocalRequest(data requestLogData, reason string) {
	slog.Info("request",
		"method", data.method, "url", data.url, "path", data.path, "proto", data.proto,
		"status", data.status, "duration_ms", data.duration.Milliseconds(),
		"ip", data.requestIP, "geo", reason,
	)
}

func fetchIPInfoContext(ctx context.Context, ip string) (*ipInfo, error) {
	if _, err := netip.ParseAddr(ip); err != nil {
		return nil, fmt.Errorf("%w: %q", errInvalidIP, ip)
	}

	token := os.Getenv("IP_INFO_TOKEN")
	if token == "" {
		return nil, errMissingIPInfoToken
	}

	requestURL := url.URL{Scheme: "https", Host: "ipinfo.io", Path: "/" + ip + "/json"}
	query := requestURL.Query()
	query.Set("token", token)
	requestURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := ipInfoHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("ipinfo returned status %d", res.StatusCode)
	}

	var info ipInfo
	if err := json.NewDecoder(io.LimitReader(res.Body, maxIPInfoResponseBytes)).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

func isInCIDR(ip, cidr string) (bool, error) {
	if cidr == "" {
		return false, nil
	}

	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return false, fmt.Errorf("%w: %q", errInvalidIP, ip)
	}
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return false, err
	}
	return prefix.Contains(addr), nil
}
