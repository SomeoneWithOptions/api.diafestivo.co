package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"time"
)

const maxIPInfoResponseBytes = 1 << 20

var (
	errMissingIPInfoToken = errors.New("missing IP_INFO_TOKEN")
	errInvalidIP          = errors.New("invalid ip address")
	ipInfoHTTPClient      = &http.Client{Timeout: 3 * time.Second}
)

type ipInfo struct {
	IP      string `json:"ip,omitempty"`
	City    string `json:"city,omitempty"`
	Region  string `json:"region,omitempty"`
	Country string `json:"country,omitempty"`
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
