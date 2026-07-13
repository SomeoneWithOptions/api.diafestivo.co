package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

const (
	contentTypeHeader = "Content-Type"
	corsHeader        = "Access-Control-Allow-Origin"
	invalidRouteBody  = `{"status":400,"message":"Please Use Valid Routes:","valid_routes":["/all","/next","/is/YYYY-MM-DD","/make?year=YYYY"]}`
)

//go:embed views/index.html
var indexTemplateSource string

//go:embed views/left.html
var leftTemplateSource string

var (
	indexTemplate = template.Must(template.New("index.html").Parse(indexTemplateSource))
	leftTemplate  = template.Must(template.New("left.html").Parse(leftTemplateSource))
	monthNames    = [...]string{"", "Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio", "Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre"}
	weekdayNames  = [...]string{"Domingo", "Lunes", "Martes", "Miércoles", "Jueves", "Viernes", "Sábado"}
)

type templateView struct {
	IsToday   bool
	Name      string
	DaysUntil int
	GifURL    *string
	Day       int
	Month     string
	Year      int
	Weekday   string
}

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

func handleAll(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, MakeHolidaysByYear(NowInCOT().Year()))
}

func handleNext(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, FindUpcomingHoliday())
}

func handleIs(w http.ResponseWriter, r *http.Request) {
	inputDate := r.PathValue("date")
	parsedDate, err := time.Parse("2006-01-02", inputDate)
	if err != nil || len(inputDate) != 10 {
		writeTextError(w, http.StatusBadRequest, "error parsing date")
		return
	}

	response := map[string]bool{"isHoliday": false}
	for _, holiday := range MakeHolidaysByYear(parsedDate.Year()) {
		if IsSameDate(parsedDate, HolidayDateInCOT(holiday)) {
			response["isHoliday"] = true
			break
		}
	}
	writeJSON(w, http.StatusOK, response)
}

func handleMake(w http.ResponseWriter, r *http.Request) {
	year, err := strconv.Atoi(r.URL.Query().Get("year"))
	if err != nil {
		writeTextError(w, http.StatusBadRequest, "error parsing year")
		return
	}
	writeJSON(w, http.StatusOK, MakeHolidaysByYear(year))
}

func handleTemplate(w http.ResponseWriter, r *http.Request) {
	var gifURL *string
	nextHoliday := FindUpcomingHoliday()
	if nextHoliday.IsToday {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		var err error
		gifURL, err = fetchGifURLContext(ctx)
		if err != nil {
			slog.Error("failed to fetch giphy gif", "error", err)
		}
	}

	holidayDate := HolidayDateInCOT(Holiday{Date: nextHoliday.Date})
	renderTemplate(w, indexTemplate, templateView{
		Name:      nextHoliday.Name,
		IsToday:   nextHoliday.IsToday,
		DaysUntil: nextHoliday.DaysUntil,
		GifURL:    gifURL,
		Day:       holidayDate.Day(),
		Month:     monthNames[int(holidayDate.Month())],
		Year:      holidayDate.Year(),
		Weekday:   weekdayNames[int(holidayDate.Weekday())],
	})
}

func handleLeft(w http.ResponseWriter, _ *http.Request) {
	type leftHolidayView struct {
		Name     string
		Day      int
		DaysLeft int
		WeekDay  string
		Month    string
	}

	year := NowInCOT().Year()
	remaining := MakeHolidaysByYear(year).GetRemaining()
	const minDaysToShow = 3
	if len(remaining) < minDaysToShow {
		nextYear := MakeHolidaysByYear(year + 1)
		remaining = append(remaining, nextYear[:minDaysToShow-len(remaining)]...)
	}

	view := make([]leftHolidayView, 0, len(remaining))
	for _, holiday := range remaining {
		date := HolidayDateInCOT(holiday)
		view = append(view, leftHolidayView{
			Name: holiday.Name, Day: date.Day(), DaysLeft: holiday.DaysUntil(),
			WeekDay: weekdayNames[int(date.Weekday())], Month: monthNames[int(date.Month())],
		})
	}
	renderTemplate(w, leftTemplate, view)
}

func handleInvalidRoute(w http.ResponseWriter, _ *http.Request) {
	setCORS(w)
	w.Header().Set(contentTypeHeader, "application/json")
	w.WriteHeader(http.StatusBadRequest)
	if _, err := w.Write([]byte(invalidRouteBody)); err != nil {
		slog.Error("failed to write json response", "error", err)
	}
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	setCORS(w)
	w.Header().Set(contentTypeHeader, "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		slog.Error("failed to write healthz response", "error", err)
	}
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

func renderTemplate(w http.ResponseWriter, tmpl *template.Template, data any) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		slog.Error("failed to execute template", "template", tmpl.Name(), "error", err)
		writeTextError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeHTML(w, http.StatusOK, buf.Bytes())
}
