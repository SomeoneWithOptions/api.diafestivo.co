package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	r "github.com/redis/go-redis/v9"
)

var redisClient *r.Client

func main() {

	os.Setenv("GODEBUG", "tls13early=1")

	godotenv.Load()

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
	http.HandleFunc("/template", HandleTemplateRoute)
	http.HandleFunc("/", HandleInvalidRoute)

	server := &http.Server{
		TLSConfig: &tls.Config{
			SessionTicketsDisabled: false,
		},
	}

	http.ListenAndServe(":"+PORT, server.Handler)
}
