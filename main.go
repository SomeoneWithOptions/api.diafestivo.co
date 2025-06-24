package main

import (
	"net/http"
	"os"
)

func main() {
	PORT := os.Getenv("PORT")

	if PORT == "" {
		PORT = "3002"
	}

	http.HandleFunc("GET /all", HandleAllRoute)
	http.HandleFunc("GET /next", HandleNextRoute)
	http.HandleFunc("GET /template", HandleTemplateRoute)
	http.HandleFunc("GET /is/{date}", HandleIsRoute)
	http.HandleFunc("GET /left", LeftHandler)
	http.HandleFunc("GET /make", MakeHandler)
	http.HandleFunc("/", HandleInvalidRoute)

	http.ListenAndServe(":"+PORT, nil)
}
