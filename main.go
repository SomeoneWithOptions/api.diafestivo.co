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
		result, err  := database.GetAllHolidaysAsJSON(REDIS_DB)
		if err != nil {
			panic (err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(*result))
		time_iso := time.Now().Format(time.RFC3339)
		fmt.Printf("the URL \"%v\"  was requested at %v\n", r.URL, time_iso)
	})
	
	http.HandleFunc("/next", func(w http.ResponseWriter, r *http.Request) {
		result, err  :=  database.GetAllHolidays(REDIS_DB)
		if err != nil {
			panic (err)
		}

		holiday.SortHolidaysArray(*result)
		fmt.Println(result)
		// w.Header().Set("Content-Type", "application/json")
		time_iso := time.Now().Format(time.RFC3339)
		message := fmt.Sprintf("Hellos World at %v", time_iso)


		w.WriteHeader(http.StatusOK)
		w.Write([]byte(message))
		fmt.Printf("the URL \"%v\"  was requested at %v\n", r.URL, time_iso)
	
	})

	fmt.Printf("listening on %s\n", PORT)
	http.ListenAndServe(":"+PORT, nil)
}
