package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/SomeoneWithOptions/api.diafestivo.co/holiday"
)

func logRequest(r *http.Request) {
	requestIP := strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]
	envCIDR := os.Getenv("MY_CIDR")

	if s, _ := IP(requestIP).IsInCIDR(envCIDR); s {
		return
	}

	p := r.Header.Get("X-Forwarded-Proto")
	t, _ := holiday.MakeDatesInCOT(holiday.Holiday{})

	ipInfo, err := IP(requestIP).FetchIPInfo()
	if err != nil {
		fmt.Println(err)
		return
	}

	slog.Info("request",
		"time", t.Format("02-01-2006:15:04:05"),
		"url", r.URL.String(),
		"proto", p,
		"ip", ipInfo.IP,
		"city", ipInfo.City,
		"region", ipInfo.Region,
		"country", ipInfo.Country,
	)
}
