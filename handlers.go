package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"
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

func HandleAllRoute(w http.ResponseWriter, r *http.Request) {
	logMessage(r)
	currentDate, _ := holiday.MakeDatesInCOT(holiday.Holiday{})
	h := holiday.MakeHolidaysByYear(currentDate.Year())

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	err := json.NewEncoder(w).Encode(h)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func HandleNextRoute(w http.ResponseWriter, r *http.Request) {
	go logMessage(r)
	n := holiday.GetNextHoliday()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	err := json.NewEncoder(w).Encode(n)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func HandleInvalidRoute(w http.ResponseWriter, r *http.Request) {
	m := InvalidRoute{400, "Please Use Valid Routes :", []string{"/all", "/next", "/is/YYYY-MM-DD"}}
	invalidRouteResponse, _ := json.Marshal(m)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(invalidRouteResponse)
}

func HandleTemplateRoute(w http.ResponseWriter, r *http.Request) {
	go logMessage(r)

	var gifURL *string
	h := holiday.GetNextHoliday()

	if h.IsToday {
		gifURL = giphy.GetGifURL()
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

	template, err := template.ParseFiles("./views/index.html")

	if err != nil {
		panic("error parsing template")
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	template.Execute(w, templateInfo)
}

func HandleIsRoute(w http.ResponseWriter, r *http.Request) {
	go logMessage(r)

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
		_, hDate := holiday.MakeDatesInCOT(h)
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

func AddClapsRoute(w http.ResponseWriter, r *http.Request) {
	go logMessage(r)
	origin := r.Header.Get("Origin")
	if !strings.Contains(origin, "diafestivo.co") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("origin not allowed"))
		return
	}
	c, _ := (redisClient.Get(r.Context(), "diafestivo:claps")).Result()
	cn, _ := strconv.Atoi(c)
	redisClient.Set(r.Context(), "diafestivo:claps", cn+1, 0)
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("👏"))
}

func GetClapsRoute(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if !strings.Contains(origin, "diafestivo.co") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("origin not allowed"))
		return
	}
	c, _ := (redisClient.Get(r.Context(), "diafestivo:claps")).Result()
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(c))
}

func LeftHandler(w http.ResponseWriter, r *http.Request) {
	go logMessage(r)

	type LeftHolidays struct {
		Name     string
		Day      int
		DaysLeft int
		WeekDay  string
		Month    string
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	template, err := template.ParseFiles("./views/left.html")
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	t, _ := holiday.MakeDatesInCOT(holiday.Holiday{})
	year := t.Year()

	allHolidays := holiday.MakeHolidaysByYear(year)
	filteredHolidays := allHolidays.FilterSundays()
	remainingHolidays := filteredHolidays.GetRemaining()

	if len(*remainingHolidays) <= 1 {
		nextYear := year + 1
		allNextYear := holiday.MakeHolidaysByYear(nextYear)
		filteredNextYear := allNextYear.FilterSundays()

		for i, a := range *filteredNextYear {
			if i == 3 {
				break
			}
			*remainingHolidays = append(*remainingHolidays, a)
		}
	}

	templateData := []LeftHolidays{}

	for _, h := range *remainingHolidays {
		_, d := holiday.MakeDatesInCOT(h)

		templateData = append(templateData, LeftHolidays{
			Name:     h.Name,
			Day:      d.Day(),
			DaysLeft: h.DaysUntil(),
			WeekDay:  weekDays[int(d.Weekday())],
			Month:    months[int(d.Month())],
		})
	}

	err = template.Execute(w, templateData)

	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

}

func MakeHandler(w http.ResponseWriter, r *http.Request) {
	go logMessage(r)

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
