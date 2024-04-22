package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	r "github.com/redis/go-redis/v9"
)

var redisClient *r.Client

func init() {
	godotenv.Load()

	opt, errParse := r.ParseURL(os.Getenv("REDIS_DB"))
	if errParse != nil {
		fmt.Printf("error parsing DB String: %v", errParse)
	}

	redisClient = r.NewClient(opt)

	err := redisClient.Ping(context.Background()).Err()

	if err != nil {
		fmt.Printf("error pinging DB: %v", err)
	}
}

func main() {
	defer redisClient.Close()

	PORT := os.Getenv("PORT")

	if PORT == "" {
		PORT = "3002"
	}

	http.HandleFunc("/all", HandleAllRoute)
	http.HandleFunc("/next", HandleNextRoute)
	http.HandleFunc("/template", HandleTemplateRoute)
	http.HandleFunc("/en", HandleEnglishRoute)
	http.HandleFunc("GET /is/{id}", HandleIsRoute)
	http.HandleFunc("POST /clap", AddClapsRoute)
	http.HandleFunc("GET /clap", GetClapsRoute)
	http.HandleFunc("/", HandleInvalidRoute)

	http.ListenAndServe(":"+PORT, nil)
}
