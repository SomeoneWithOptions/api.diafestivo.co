package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		fmt.Printf("error loading .env file: %v\n", err)
	}

	PORT := os.Getenv("PORT")

	if PORT == "" {
		PORT = "3002"
	}

	http.HandleFunc("/all", HandleAllRoute)
	http.HandleFunc("/next", HandleNextRoute)
	http.HandleFunc("/gif", HandleGifRoute)
	http.HandleFunc("/", HandleInvaliedRoute)

	fmt.Println("running at", PORT)
	http.ListenAndServe(":"+PORT, nil)
}
