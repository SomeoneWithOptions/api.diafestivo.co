package main

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/SomeoneWithOptions/api.diafestivo.co/giphy"
	"github.com/SomeoneWithOptions/api.diafestivo.co/holiday"
	"github.com/SomeoneWithOptions/api.diafestivo.co/templateinfo"
)

type InvalidRoute struct {
	Status      int      `json:"status"`
	Message     string   `json:"message"`
	ValidRoutes []string `json:"valid_routes"`
}

var invalidRouteResponse []byte

// Pre-parsed templates — avoids disk I/O + parsing on every request.
var (
	indexTemplate = template.Must(template.ParseFiles("./views/index.html"))
	leftTemplate  = template.Must(template.ParseFiles("./views/left.html"))
)

var months = map[int]string{
	1:  "Enero",
	2:  "Febrero",
	3:  "Marzo",
	4:  "Abril",
	5:  "Mayo",
	6:  "Junio",
	7:  "Julio",
	8:  "Agosto",
	9:  "Septiembre",
	10: "Octubre",
	11: "Noviembre",
	12: "Diciembre",
}

var weekDays = map[int]string{
	1: "Lunes",
	2: "Martes",
	3: "Míercoles",
	4: "Jueves",
	5: "Viernes",
	6: "Sábado",
	0: "Domingo",
}

func init() {
	m := InvalidRoute{
		Status:  http.StatusBadRequest,
		Message: "Please Use Valid Routes:",
		ValidRoutes: []string{
			"/all",
			"/next",
			"/is/YYYY-MM-DD",
			"/make?year=YYYY",
		},
	}
	var err error
	invalidRouteResponse, err = json.Marshal(m)
	if err != nil {
		slog.Error("failed to marshal invalid route response", "error", err)
		os.Exit(1)
	}
}

func HandleAllRoute(w http.ResponseWriter, r *http.Request) {
	go logRequest(r)
	now := holiday.NowInCOT()
	h := holiday.MakeHolidaysByYear(now.Year())

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := json.NewEncoder(w).Encode(h); err != nil {
		slog.Error("failed to encode all route response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func HandleNextRoute(w http.ResponseWriter, r *http.Request) {
	go logRequest(r)
	n := holiday.FindUpcomingHoliday()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := json.NewEncoder(w).Encode(n); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func HandleTemplateRoute(w http.ResponseWriter, r *http.Request) {
	go logRequest(r)

	var gifURL *string
	h := holiday.FindUpcomingHoliday()

	if h.IsToday {
		gifURL = giphy.FetchGifURL()
	}

	templateInfo := templateinfo.NewTemplateInfo(
		h.Name,
		h.IsToday,
		h.DaysUntil,
		h.Date,
		gifURL,
		h.Date.Day(),
		months[int(h.Date.Month())],
		h.Date.Year(),
		weekDays[int(h.Date.Weekday())],
	)

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	indexTemplate.Execute(w, templateInfo)
}

func HandleIsRoute(w http.ResponseWriter, r *http.Request) {
	go logRequest(r)

	response := make(map[string]bool)
	inputDate := r.PathValue("date")
	layout := "2006-01-02"
	t, err := time.Parse(layout, inputDate)

	if err != nil || len(inputDate) != 10 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error parsing date"))
		return
	}

	allHolidays := holiday.MakeHolidaysByYear(t.Year())

	for _, h := range *allHolidays {
		hDate := holiday.HolidayDateInCOT(h)
		if holiday.IsSameDate(t, hDate) {
			response["isHoliday"] = true
			break
		}
	}

	if !response["isHoliday"] {
		response["isHoliday"] = false
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func LeftHandler(w http.ResponseWriter, r *http.Request) {
	go logRequest(r)

	type LeftHolidays struct {
		Name     string
		Day      int
		DaysLeft int
		WeekDay  string
		Month    string
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	now := holiday.NowInCOT()
	year := now.Year()

	allHolidays := holiday.MakeHolidaysByYear(year)
	remainingHolidays := allHolidays.GetRemaining()

	minDaystoShow := 3

	if len(*remainingHolidays) < minDaystoShow {
		allNextYear := holiday.MakeHolidaysByYear(year + 1)
		n := minDaystoShow - len(*remainingHolidays)
		for i, a := range *allNextYear {
			if i == n {
				break
			}
			*remainingHolidays = append(*remainingHolidays, a)
		}
	}

	templateData := []LeftHolidays{}

	for _, h := range *remainingHolidays {
		d := holiday.HolidayDateInCOT(h)

		templateData = append(templateData, LeftHolidays{
			Name:     h.Name,
			Day:      d.Day(),
			DaysLeft: h.DaysUntil(),
			WeekDay:  weekDays[int(d.Weekday())],
			Month:    months[int(d.Month())],
		})
	}

	err := leftTemplate.Execute(w, templateData)

	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

}

func MakeHandler(w http.ResponseWriter, r *http.Request) {
	go logRequest(r)

	queryParams := r.URL.Query()

	yearInput := queryParams.Get("year")
	year, err := strconv.Atoi(yearInput)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error parsing year"))
		return
	}

	h := holiday.MakeHolidaysByYear(year)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	err = json.NewEncoder(w).Encode(h)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func HandleInvalidRoute(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(invalidRouteResponse)
}
