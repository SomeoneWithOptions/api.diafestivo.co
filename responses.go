package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

const (
	contentTypeHeader = "Content-Type"
	corsHeader        = "Access-Control-Allow-Origin"
)

type invalidRouteResponse struct {
	Status      int      `json:"status"`
	Message     string   `json:"message"`
	ValidRoutes []string `json:"valid_routes"`
}

var invalidRouteResponseBody []byte

func init() {
	body := invalidRouteResponse{
		Status:  http.StatusBadRequest,
		Message: "Please Use Valid Routes:",
		ValidRoutes: []string{
			"/all",
			"/next",
			"/is/YYYY-MM-DD",
			"/make?year=YYYY",
		},
	}

	encodedBody, err := json.Marshal(body)
	if err != nil {
		slog.Error("failed to marshal invalid route response", "error", err)
		os.Exit(1)
	}
	invalidRouteResponseBody = encodedBody
}

func setCORS(w http.ResponseWriter) {
	w.Header().Set(corsHeader, "*")
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	setCORS(w)
	w.Header().Set(contentTypeHeader, "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("failed to encode json response", "error", err)
	}
}

func writeJSONBytes(w http.ResponseWriter, status int, body []byte) {
	setCORS(w)
	w.Header().Set(contentTypeHeader, "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		slog.Error("failed to write json response", "error", err)
	}
}

func writeTextError(w http.ResponseWriter, status int, body string) {
	setCORS(w)
	w.WriteHeader(status)
	if _, err := w.Write([]byte(body)); err != nil {
		slog.Error("failed to write text error response", "error", err)
	}
}

func writeHTML(w http.ResponseWriter, status int, body []byte) {
	setCORS(w)
	w.Header().Set(contentTypeHeader, "text/html")
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		slog.Error("failed to write html response", "error", err)
	}
}
