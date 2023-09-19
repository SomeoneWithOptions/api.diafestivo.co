package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Message struct {
	Text string `json:"message"`
}

func main() {

	if err := godotenv.Load(); err != nil {
		fmt.Printf("Error loading .env file: %v", err)
	}

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "3002"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		message := Message{Text: "hello"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(message)
		time_iso := time.Now().Format("2006-01-02")
		fmt.Println(time_iso)
	})

	fmt.Printf("listening on %s\n", PORT)
	http.ListenAndServe(":"+PORT, nil)
}
