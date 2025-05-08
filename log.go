package main

import (
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/SomeoneWithOptions/api.diafestivo.co/holiday"
)

func logMessage(r *http.Request) {

	var clientIP IP
	reqiestIP := strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]
	envIPs := os.Getenv("MY_IP")

	whiteListIPs := strings.Split(envIPs, ",")

	if slices.Contains(whiteListIPs, reqiestIP) {
		return
	}

	p := r.Header.Get("X-Forwarded-Proto")
	t, _ := holiday.MakeDatesInCOT(holiday.Holiday{})

	ipinfo, err := IP(reqiestIP).FetchIPInfo()
	if err != nil {
		fmt.Println(err)
		return
	}

	message := fmt.Sprintf(
		"\"%v\" %v %v %v %v %v  %v\n",
		r.URL,
		t.Format("02-01-2006:15:04:05"),
		p,
		clientIP,
		ipinfo.City,
		ipinfo.Region,
		ipinfo.Country,
	)
	fmt.Printf("%v", message)
}
