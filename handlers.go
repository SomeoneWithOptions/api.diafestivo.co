package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/SomeoneWithOptions/api.diafestivo.co/database"
	"github.com/SomeoneWithOptions/api.diafestivo.co/giphy"
	"github.com/SomeoneWithOptions/api.diafestivo.co/holiday"

	"github.com/ipinfo/go/v2/ipinfo"
)

type InvalidRoute struct {
	Status      int      `json:"status"`
	Message     string   `json:"message"`
	ValidRoutes []string `json:"valid_routes"`
}

func HandleAllRoute(w http.ResponseWriter, r *http.Request) {
	go logMessage(r)
	result, err := database.GetAllHolidaysAsJSON(redisClient)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(*result))
}

func HandleNextRoute(w http.ResponseWriter, r *http.Request) {
	go logMessage(r)
	n := MakeNextNewHoliday()
	message, _ := json.Marshal(n)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(message))
}

func HandleGifRoute(w http.ResponseWriter, r *http.Request) {
	go logMessage(r)
	gif_url := giphy.GetGifURL()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(gif_url))
}

func HandleInvalidRoute(w http.ResponseWriter, r *http.Request) {
	go logMessage(r)
	m := InvalidRoute{404, "Please Use Valid Routes :", []string{"/all", "/next"}}
	invalidRouteResponse, _ := json.Marshal(m)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(invalidRouteResponse)
}

func HandleTemplateRoute(w http.ResponseWriter, r *http.Request) {
	go logMessage(r)
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	nh := MakeNextNewHoliday()
	n := holiday.NewNextHoliday(nh.Name, nh.Date, false, 2)
	tmpl, _ := template.ParseFiles("./template/index.html")
	tmpl.Execute(w, n)
}

func MakeNextNewHoliday() holiday.NextHoliday {
	current_year := time.Now().Year()
	var all_holidays *[]holiday.Holiday
	var err error
	all_holidays, err = database.GetAllHolidays(redisClient, current_year)
	if err != nil {
		panic(err)
	}

	holiday.SortHolidaysArray(*all_holidays)
	var next_holiday = holiday.FindNextHoliday(*all_holidays)

	if next_holiday == nil {
		next_year := time.Now().Year() + 1
		all_holidays, _ = database.GetAllHolidays(redisClient, next_year)
		holiday.SortHolidaysArray(*all_holidays)
		next_holiday = holiday.FindNextHoliday(*all_holidays)
	}

	n := holiday.NewNextHoliday(next_holiday.Name, next_holiday.Date, next_holiday.IsToday(), int32(next_holiday.DaysUntil()))
	return n
}

func logMessage(r *http.Request) {
	token := os.Getenv("IP_INFO_TOKEN")
	ip := strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]
	p := r.Header.Get("X-Forwarded-Proto")
	t, _ := holiday.MakeDates(holiday.Holiday{})
	ip_info_client := ipinfo.NewClient(nil, nil, token)
	info, _ := ip_info_client.GetIPInfo(net.ParseIP(ip))
	fmt.Printf("\"%v\" %v %v %v %v %v %v\n", r.URL, t.Format("02-01-2006:15:04:05"), p, ip, info.City, info.Region, info.Country)
}
