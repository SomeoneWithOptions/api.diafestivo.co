package main

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/SomeoneWithOptions/api.diafestivo.co/holiday"
)

func handleAll(w http.ResponseWriter, _ *http.Request) {
	now := holiday.NowInCOT()
	holidays := holiday.MakeHolidaysByYear(now.Year())
	writeJSON(w, http.StatusOK, holidays)
}

func handleNext(w http.ResponseWriter, _ *http.Request) {
	nextHoliday := holiday.FindUpcomingHoliday()
	writeJSON(w, http.StatusOK, nextHoliday)
}

func handleIs(w http.ResponseWriter, r *http.Request) {
	inputDate := r.PathValue("date")
	parsedDate, err := time.Parse("2006-01-02", inputDate)
	if err != nil || len(inputDate) != 10 {
		writeTextError(w, http.StatusBadRequest, "error parsing date")
		return
	}

	response := map[string]bool{"isHoliday": false}
	holidays := holiday.MakeHolidaysByYear(parsedDate.Year())
	for _, holidayItem := range *holidays {
		holidayDate := holiday.HolidayDateInCOT(holidayItem)
		if holiday.IsSameDate(parsedDate, holidayDate) {
			response["isHoliday"] = true
			break
		}
	}

	writeJSON(w, http.StatusOK, response)
}

func handleMake(w http.ResponseWriter, r *http.Request) {
	yearInput := r.URL.Query().Get("year")
	year, err := strconv.Atoi(yearInput)
	if err != nil {
		writeTextError(w, http.StatusBadRequest, "error parsing year")
		return
	}

	holidays := holiday.MakeHolidaysByYear(year)
	writeJSON(w, http.StatusOK, holidays)
}

func handleInvalidRoute(w http.ResponseWriter, _ *http.Request) {
	writeJSONBytes(w, http.StatusBadRequest, invalidRouteResponseBody)
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	setCORS(w)
	w.Header().Set(contentTypeHeader, "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		slog.Error("failed to write healthz response", "error", err)
	}
}
