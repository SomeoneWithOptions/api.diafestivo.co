package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/SomeoneWithOptions/api.diafestivo.co/holiday"
)

func logMessage(r *http.Request) {
	requestIP := strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]
	envIPs := os.Getenv("MY_IP")

	whiteListIPs := strings.Split(envIPs, ",")

	if slices.Contains(whiteListIPs, requestIP) {
		return
	}

	p := r.Header.Get("X-Forwarded-Proto")
	t, _ := holiday.MakeDatesInCOT(holiday.Holiday{})

	ipinfo, err := IP(requestIP).FetchIPInfo()
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Printf("[NOTICE] \"%v\" %v %v %v %v %v %v\n",
		r.URL,
		t.Format("02-01-2006:15:04:05"),
		p,
		ipinfo.IP,
		ipinfo.City,
		ipinfo.Region,
		ipinfo.Country)
}
