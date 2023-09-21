package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

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
		time_iso := time.Now().Format(time.RFC3339)
		fmt.Printf("the URL \"%v\"  was requested at %v\n", r.URL, time_iso)
		result, err := database.GetAllHolidaysAsJSON(REDIS_DB)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(*result))
	})

	http.HandleFunc("/next", func(w http.ResponseWriter, r *http.Request) {

		fmt.Printf("local: %v -- utc-5:%v\n", time.Now().Format(time.RFC3339), holiday.GetUTC5Time().Format(time.RFC3339))
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
	fmt.Println("running at", PORT)
	http.ListenAndServe(":"+PORT, nil)

}

func HandleGifRoute(w http.ResponseWriter, r *http.Request) {
	gif_url := giphy.GetGifURL()
	w.Write([]byte(gif_url))
}
