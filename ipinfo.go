package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
)

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
	Asn           string `json:"asn,omitempty"`
	AsName        string `json:"as_name,omitempty"`
	AsDomain      string `json:"as_domain,omitempty"`
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
	var ipinfo IPInfo
	token := os.Getenv("IP_INFO_TOKEN")
	url := fmt.Sprintf("https://ipinfo.io/%s/json?token=%s", ip, token)

	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	err = json.NewDecoder(r.Body).Decode(&ipinfo)
	if err != nil {
		return nil, err
	}
	return &ipinfo, nil
}

func (ip IP) FetchIPInfoLite() (*IPInfoLite, error) {
	var ipinfo IPInfoLite
	token := os.Getenv("IP_INFO_TOKEN")
	url := fmt.Sprintf("https://api.ipinfo.io/lite/%s/json?token=%s", ip, token)

	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	err = json.NewDecoder(r.Body).Decode(&ipinfo)
	if err != nil {
		return nil, err
	}
	return &ipinfo, nil
}

func (i IP) IsInCIDR(cidr string) (bool, error) {
	parsedIP := net.ParseIP(string(i))
	_, parsedNet, err := net.ParseCIDR(cidr)

	if err != nil {
		return false, err
	}

	return parsedNet.Contains(parsedIP), nil

}
