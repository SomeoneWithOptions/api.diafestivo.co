package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/SomeoneWithOptions/api.diafestivo.co/database"
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
		fmt.Printf("the URL \"%v\"  was requested at %v\n", r.URL, time.Now().Format(time.RFC3339))
		all_holidays, err := database.GetAllHolidays(REDIS_DB)
		if err != nil {
			panic(err)
		}

		holiday.SortHolidaysArray(*all_holidays)
		next_holiday := holiday.FindNextHoliday(*all_holidays)
		fmt.Println(next_holiday)
		// w.Header().Set("Content-Type", "application/json")
		message := fmt.Sprintf("el Siguiente Festivo es %v %v", next_holiday.Name, next_holiday.Date)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(message))

	})

	fmt.Printf("listening on %s\n", PORT)
	http.ListenAndServe(":"+PORT, nil)
}
