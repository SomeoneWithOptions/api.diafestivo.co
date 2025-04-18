package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/SomeoneWithOptions/api.diafestivo.co/giphy"
	"github.com/SomeoneWithOptions/api.diafestivo.co/holiday"
	"github.com/SomeoneWithOptions/api.diafestivo.co/templateinfo"

	"github.com/ipinfo/go/v2/ipinfo"
	j "github.com/json-iterator/go"
)

type InvalidRoute struct {
	Status      int      `json:"status"`
	Message     string   `json:"message"`
	ValidRoutes []string `json:"valid_routes"`
}

type IsHoliday struct {
	IsHoliday bool `json:"is_holiday"`
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
	jsonResponse, _ := j.Marshal(h)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func HandleNextRoute(w http.ResponseWriter, r *http.Request) {
	go logMessage(r)
	defer r.Body.Close()
	n := GetNextHoliday()
	n_holiday_json, _ := j.Marshal(n)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(n_holiday_json))
}

func HandleInvalidRoute(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	m := InvalidRoute{400, "Please Use Valid Routes :", []string{"/all", "/next", "/is/YYYY-MM-DD"}}
	invalidRouteResponse, _ := json.Marshal(m)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(invalidRouteResponse)
}

func HandleTemplateRoute(w http.ResponseWriter, r *http.Request) {
	go logMessage(r)
	defer r.Body.Close()
	var gif_url *string
	nh := GetNextHoliday()

	if nh.IsToday {
		gif_url = giphy.GetGifURL()
	}

	t_info := templateinfo.NewTemplateInfo(
		nh.Name,
		nh.IsToday,
		nh.DaysUntil,
		nh.Date,
		gif_url,
		nh.Date.Day(),
		months[int(nh.Date.Month())],
		nh.Date.Year(),
		weekDays[int(nh.Date.Weekday())],
	)

	tmpl, err := template.ParseFiles("./views/index.html")

	if err != nil {
		panic("error parsing template")
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, t_info)
}

func GetNextHoliday() *holiday.NextHoliday {
	c_date, _ := holiday.MakeDatesInCOT(holiday.Holiday{})
	
	a_holidays := holiday.MakeHolidaysByYear(c_date.Year())

	holiday.SortHolidaysArray(*a_holidays)
	var n_holiday = holiday.FindNextHoliday(*a_holidays)

	if n_holiday == nil {
		next_year := c_date.Year() + 1
		a_holidays:= holiday.MakeHolidaysByYear(next_year)
		n_holiday = holiday.FindNextHoliday(*a_holidays)
	}

	n := holiday.NewNextHoliday(
		n_holiday.Name,
		n_holiday.Date,
		n_holiday.IsToday(),
		n_holiday.DaysUntil(),
	)
	return &n
}

func HandleIsRoute(w http.ResponseWriter, r *http.Request) {
	go logMessage(r)
	defer r.Body.Close()

	inputDate := r.PathValue("id")

	if len(inputDate) != 10 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error parsing date"))
		return
	}

	layout := "2006-01-02"
	t, err := time.Parse(layout, inputDate)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error parsing date"))
		return
	}

	allHolidays := holiday.MakeHolidaysByYear(t.Year())

	for _, h := range *allHolidays {
		_, hDate := holiday.MakeDatesInCOT(h)
		is := holiday.IsSameDate(t, hDate)
		if is {
			res := IsHoliday{true}
			g, _ := j.Marshal(res)

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.WriteHeader(http.StatusOK)
			w.Write(g)
			return
		}
	}

	res := IsHoliday{false}
	g, _ := j.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(g)
}

func AddClapsRoute(w http.ResponseWriter, r *http.Request) {
	go logMessage(r)
	defer r.Body.Close()
	c, _ := (redisClient.Get(r.Context(), "diafestivo:claps")).Result()
	cn, _ := strconv.Atoi(c)
	redisClient.Set(r.Context(), "diafestivo:claps", cn+1, 0)
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("👏"))
}

func GetClapsRoute(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	c, _ := (redisClient.Get(r.Context(), "diafestivo:claps")).Result()
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(c))
}

func LeftHandler(w http.ResponseWriter, r *http.Request) {
	go logMessage(r)
	defer r.Body.Close()

	type LeftHolidays struct {
		Name     string
		Day      int
		DaysLeft int
		WeekDay  string
		Month    string
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	tmpl, err := template.ParseFiles("./views/left.html")
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	t, _ := holiday.MakeDatesInCOT(holiday.Holiday{})
	year := t.Year()

	all := holiday.MakeHolidaysByYear(year)
	remaining := holiday.GetRemainingHolidaysInYear(all, year)

	if len(*remaining) <= 1 {
		nextYear := year + 1
		allNextYear := holiday.MakeHolidaysByYear(nextYear)

		for i, a := range *allNextYear {
			if i == 3 {
				break
			}
			*remaining = append(*remaining, a)
		}
	}

	data := []LeftHolidays{}

	for _, h := range *remaining {
		_, d := holiday.MakeDatesInCOT(h)

		data = append(data, LeftHolidays{
			Name:     h.Name,
			Day:      d.Day(),
			DaysLeft: h.DaysUntil(),
			WeekDay:  weekDays[int(d.Weekday())],
			Month:    months[int(d.Month())],
		})
	}

	err = tmpl.Execute(w, data)

	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

}

func MakeHandler(w http.ResponseWriter, r *http.Request) {
	go logMessage(r)
	defer r.Body.Close()
	queryParams := r.URL.Query()

	yearInput := queryParams.Get("year")

	year, err := strconv.Atoi(yearInput)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error parsing year"))
		return
	}

	h := holiday.MakeHolidaysByYear(year)
	json, _ := json.Marshal(h)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func logMessage(r *http.Request) {

	token := os.Getenv("IP_INFO_TOKEN")
	clientIP := strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]
	envIPs := os.Getenv("MY_IP")

	whiteListIPs := strings.Split(envIPs, ",")

	if slices.Contains(whiteListIPs, clientIP) {
		return
	}

	p := r.Header.Get("X-Forwarded-Proto")
	t, _ := holiday.MakeDatesInCOT(holiday.Holiday{})

	ipInfoClient := ipinfo.NewClient(nil, nil, token)
	info, err := ipInfoClient.GetIPInfo(net.ParseIP(clientIP))

	if err != nil {
		info = &ipinfo.Core{City: "NO IP INFO"}
	}

	message := fmt.Sprintf(
		"\"%v\" %v %v %v %v %v  %v\n",
		r.URL,
		t.Format("02-01-2006:15:04:05"),
		p,
		clientIP,
		info.City,
		info.Region,
		info.Country,
	)
	fmt.Printf("%v", message)
}
