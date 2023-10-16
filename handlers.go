package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/SomeoneWithOptions/api.diafestivo.co/database"
	"github.com/SomeoneWithOptions/api.diafestivo.co/giphy"
	"github.com/SomeoneWithOptions/api.diafestivo.co/holiday"
)

type InvalidRoute struct {
	Status      int      `json:"status"`
	Message     string   `json:"message"`
	ValidRoutes []string `json:"valid_routes"`
}

func HandleAllRoute(w http.ResponseWriter, r *http.Request) {
	logMessage(r)
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
	logMessage(r)
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
	message, _ := json.Marshal(n)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(message))
}

func HandleGifRoute(w http.ResponseWriter, r *http.Request) {
	logMessage(r)
	gif_url := giphy.GetGifURL()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(gif_url))
}

func HandleInvaliedRoute(w http.ResponseWriter, r *http.Request) {
	logMessage(r)
	m := InvalidRoute{404, "Please Use Valid Routes :", []string{"/all", "/next"}}
	invalidRouteResponse, _ := json.Marshal(m)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(invalidRouteResponse)
}

func logMessage(r *http.Request) {
	ip := r.Header.Get("X-Forwarded-For")
	p := r.Header.Get("X-Forwarded-Proto")
	t, _ := holiday.MakeDates(holiday.Holiday{})
	fmt.Printf("requested: \"%v\" at: %v from: %v over: %v\n", r.URL, t.Format("02-01-2006:15:04:05"), ip, p)
}
