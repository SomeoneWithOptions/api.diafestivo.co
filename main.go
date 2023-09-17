package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/SomeoneWithOptions/api.diafestivo.co/holiday"
)

type Message struct {
	Text string `json:"message"`
}

func main() {
	PORT := "3002"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		message := Message{Text: "hello"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(message)
		current_time := time.Now()
		fmt.Println(current_time)
	})

	cristmas_holiday := holiday.NewNextHoliday("Christmas", "31/12/2023", false, 200)
	fmt.Println(cristmas_holiday.Print())
	http.ListenAndServe(":"+PORT, nil)
}
