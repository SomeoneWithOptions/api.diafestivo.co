package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/SomeoneWithOptions/api.diafestivo.co/database"
	"github.com/SomeoneWithOptions/api.diafestivo.co/giphy"
	"github.com/SomeoneWithOptions/api.diafestivo.co/holiday"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		fmt.Printf("error loading .env file: %v\n", err)
	}

	PORT := os.Getenv("PORT")
	REDIS_DB := os.Getenv("REDIS_DB")

	if PORT == "" {
		PORT = "3002"
	}

	http.HandleFunc("/all", func(w http.ResponseWriter, r *http.Request) {
		t, _ := holiday.MakeDates(holiday.Holiday{})
		fmt.Printf("the URL \"%v\" was requested at %v\n", r.URL, t)
		result, err := database.GetAllHolidaysAsJSON(REDIS_DB)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(*result))
	})

	http.HandleFunc("/next", func(w http.ResponseWriter, r *http.Request) {
		t, _ := holiday.MakeDates(holiday.Holiday{})
		fmt.Printf("the URL \"%v\" was requested at %v\n", r.URL, t)
		all_holidays, err := database.GetAllHolidays(REDIS_DB)
		if err != nil {
			panic(err)
		}

		holiday.SortHolidaysArray(*all_holidays)
		next_holiday := holiday.FindNextHoliday(*all_holidays)
		n := holiday.NewNextHoliday(next_holiday.Name, next_holiday.Date, next_holiday.IsToday(), int32(next_holiday.DaysUntil()))
		message, _ := json.Marshal(n)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(message))

	})

	http.HandleFunc("/gif", HandleGifRoute)
	http.HandleFunc("/", HandleInvaliedRoute)
	fmt.Println("running at", PORT)
	http.ListenAndServe(":"+PORT, nil)
}

func HandleGifRoute(w http.ResponseWriter, r *http.Request) {
	t, _ := holiday.MakeDates(holiday.Holiday{})
	fmt.Printf("the URL \"%v\" was requested at %v\n", r.URL, t)
	gif_url := giphy.GetGifURL()
	w.Write([]byte(gif_url))
}

func HandleInvaliedRoute(w http.ResponseWriter, r *http.Request) {
	t, _ := holiday.MakeDates(holiday.Holiday{})
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 not found"))
	fmt.Printf("invalid route \"%v\" at %v\n", r.URL, t)
}