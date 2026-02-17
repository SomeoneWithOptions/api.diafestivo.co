package main

import (
	"log/slog"
	"net/http"
	"os"
)

func main() {
	handlerOptions := &slog.HandlerOptions{Level: slog.LevelInfo}
	stdoutHandler := slog.NewJSONHandler(os.Stdout, handlerOptions)
	stderrHandler := slog.NewJSONHandler(os.Stderr, handlerOptions)
	slog.SetDefault(slog.New(newSplitHandler(stdoutHandler, stderrHandler, slog.LevelError)))

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

	if err := http.ListenAndServe(":"+PORT, nil); err != nil {
		slog.Error("http server stopped", "error", err)
	}
}
