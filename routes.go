package main

import "net/http"

func newServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /all", handleAll)
	mux.HandleFunc("GET /next", handleNext)
	mux.HandleFunc("GET /template", handleTemplate)
	mux.HandleFunc("GET /is/{date}", handleIs)
	mux.HandleFunc("GET /left", handleLeft)
	mux.HandleFunc("GET /make", handleMake)
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("/", handleInvalidRoute)
	return mux
}

// Exported wrappers kept for compatibility with external tests/importers.
func HandleAllRoute(w http.ResponseWriter, r *http.Request)      { handleAll(w, r) }
func HandleNextRoute(w http.ResponseWriter, r *http.Request)     { handleNext(w, r) }
func HandleTemplateRoute(w http.ResponseWriter, r *http.Request) { handleTemplate(w, r) }
func HandleIsRoute(w http.ResponseWriter, r *http.Request)       { handleIs(w, r) }
func LeftHandler(w http.ResponseWriter, r *http.Request)         { handleLeft(w, r) }
func MakeHandler(w http.ResponseWriter, r *http.Request)         { handleMake(w, r) }
func HandleInvalidRoute(w http.ResponseWriter, r *http.Request)  { handleInvalidRoute(w, r) }
