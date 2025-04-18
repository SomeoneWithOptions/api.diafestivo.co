package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	r "github.com/redis/go-redis/v9"
)

var redisClient *r.Client

func init() {

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
	http.HandleFunc("GET /is/{id}", HandleIsRoute)
	http.HandleFunc("POST /clap", AddClapsRoute)
	http.HandleFunc("GET /clap", GetClapsRoute)
	http.HandleFunc("GET /left", LeftHandler)
	http.HandleFunc("GET /make", MakeHandler)
	http.HandleFunc("/", HandleInvalidRoute)

	http.ListenAndServe(":"+PORT, nil)
}
