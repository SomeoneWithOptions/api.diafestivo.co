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
	ErrMissingIPInfoToken = errors.New("missing IP_INFO_TOKEN")
	ErrInvalidIP          = errors.New("invalid ip address")
)

var ipInfoHTTPClient = &http.Client{Timeout: 3 * time.Second}

type IPInfo struct {
	IP       string `json:"ip,omitempty"`
	City     string `json:"city,omitempty"`
	Region   string `json:"region,omitempty"`
	Country  string `json:"country,omitempty"`
	Loc      string `json:"loc,omitempty"`
	Org      string `json:"org,omitempty"`
	Postal   string `json:"postal,omitempty"`
	Timezone string `json:"timezone,omitempty"`
	Readme   string `json:"readme,omitempty"`
}

type IPInfoLite struct {
	IP            string `json:"ip,omitempty"`
	ASN           string `json:"asn,omitempty"`
	ASName        string `json:"as_name,omitempty"`
	ASDomain      string `json:"as_domain,omitempty"`
	CountryCode   string `json:"country_code,omitempty"`
	Country       string `json:"country,omitempty"`
	ContinentCode string `json:"continent_code,omitempty"`
	Continent     string `json:"continent,omitempty"`
}

type IP string

func (ip IP) String() string {
	return string(ip)
}

func (ip IP) FetchIPInfo() (*IPInfo, error) {
	return ip.FetchIPInfoContext(context.Background())
}

func (ip IP) FetchIPInfoContext(ctx context.Context) (*IPInfo, error) {
	var ipInfo IPInfo
	if err := fetchIPInfo(ctx, "ipinfo.io", "", ip.String(), &ipInfo); err != nil {
		return nil, err
	}
	return &ipInfo, nil
}

func (ip IP) FetchIPInfoLite() (*IPInfoLite, error) {
	return ip.FetchIPInfoLiteContext(context.Background())
}

func (ip IP) FetchIPInfoLiteContext(ctx context.Context) (*IPInfoLite, error) {
	var ipInfo IPInfoLite
	if err := fetchIPInfo(ctx, "api.ipinfo.io", "/lite", ip.String(), &ipInfo); err != nil {
		return nil, err
	}
	return &ipInfo, nil
}

func fetchIPInfo(ctx context.Context, host, basePath, ip string, target any) error {
	if _, err := netip.ParseAddr(ip); err != nil {
		return fmt.Errorf("%w: %q", ErrInvalidIP, ip)
	}

	token := os.Getenv("IP_INFO_TOKEN")
	if token == "" {
		return ErrMissingIPInfoToken
	}

	requestURL := url.URL{
		Scheme: "https",
		Host:   host,
		Path:   basePath + "/" + ip + "/json",
	}
	query := requestURL.Query()
	query.Set("token", token)
	requestURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return err
	}

	res, err := ipInfoHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("ipinfo returned status %d", res.StatusCode)
	}

	if err := json.NewDecoder(io.LimitReader(res.Body, maxIPInfoResponseBytes)).Decode(target); err != nil {
		return err
	}
	return nil
}

func (ip IP) IsInCIDR(cidr string) (bool, error) {
	if cidr == "" {
		return false, nil
	}

	addr, err := netip.ParseAddr(ip.String())
	if err != nil {
		return false, fmt.Errorf("%w: %q", ErrInvalidIP, ip.String())
	}

	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return false, err
	}

	return prefix.Contains(addr), nil
}
