package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	r "github.com/redis/go-redis/v9"
)

var redisClient *r.Client

func main() {

	if err := godotenv.Load(); err != nil {
		fmt.Printf("error loading .env file: %v\n", err)
	}

	opt, errParse := r.ParseURL(os.Getenv("REDIS_DB"))
	if errParse != nil {
		fmt.Printf("error parsing DB String: %v", errParse)
	}
	redisClient = r.NewClient(opt)

	defer redisClient.Close()

	PORT := os.Getenv("PORT")

	if PORT == "" {
		PORT = "3002"
	}

	http.HandleFunc("/all", HandleAllRoute)
	http.HandleFunc("/next", HandleNextRoute)
	http.HandleFunc("/gif", HandleGifRoute)
	http.HandleFunc("/ping", HandlePingRoute)
	http.HandleFunc("/", HandleInvalidRoute)

	fmt.Println("running at", PORT)
	http.ListenAndServe(":"+PORT, nil)
}
