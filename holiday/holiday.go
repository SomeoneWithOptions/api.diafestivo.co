package holiday

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type NextHoliday struct {
	Name      string `json:"name"`
	Date      string `json:"date"`
	IsToday   bool   `json:"isToday"`
	DaysUntil int    `json:"daysUntil"`
}

type Holiday struct {
	Name string `json:"name"`
	Date string `json:"date"`
}

func NewNextHoliday(name string, date string, is_today bool, days_until int) NextHoliday {
	var next_holiday NextHoliday
	next_holiday.Name = name
	next_holiday.Date = date
	next_holiday.IsToday = is_today
	next_holiday.DaysUntil = days_until
	return next_holiday
}

func (n NextHoliday) Print() string {
	return fmt.Sprintf("name: %s\ndate: %s\nisToday: %v\ndaysUntil: %d", n.Name, n.Date, n.IsToday, n.DaysUntil)
}

func SortHolidaysArray(holidays []Holiday) {
	sort.SliceStable(holidays, func(i, j int) bool {
		dateI, _ := time.Parse(time.RFC3339, holidays[i].Date)
		dateJ, _ := time.Parse(time.RFC3339, holidays[j].Date)
		return dateI.Before(dateJ)
	})
}

func FindNextHoliday(holidays []Holiday) *Holiday {
	current_time, _ := MakeDates(Holiday{})
	for _, h := range holidays {
		_, holiday_date := MakeDates(h)

		if h.IsToday() {
			return &h
		}
		if holiday_date.After(current_time) {
			return &h
		}
	}
	return nil
}

func GetRemainingHolidaysInYear(h *[]Holiday, year int) *[]Holiday {
	var remainingHolidays []Holiday
	today, _ := MakeDates(Holiday{})

	for _, holiday := range *h {
		_, holidayDate := MakeDates(holiday)
		if holidayDate.After(today) {
			remainingHolidays = append(remainingHolidays, holiday)
		}
	}
	return &remainingHolidays
}

func (h Holiday) IsToday() bool {
	currentDate, holidayDate := MakeDates(h)
	return holidayDate.Year() == currentDate.Year() &&
		holidayDate.Month() == currentDate.Month() &&
		holidayDate.Day() == currentDate.Day()
}

func (h Holiday) DaysUntil() int {
	currentDate, holidayDate := MakeDates(h)
	daysUntil := math.Ceil(holidayDate.Sub(currentDate).Hours() / 24)
	return int(daysUntil)
}

func MakeDates(h Holiday) (time.Time, time.Time) {
	loc := time.FixedZone("UTC-5", -5*60*60)
	hd, _ := time.Parse(time.RFC3339, h.Date)
	holidayDate := hd.In(loc).Add(time.Hour * 5)
	currentDate := time.Now().In(loc)
	return currentDate, holidayDate
}

func IsSameDate(d1, d2 time.Time) bool {
	return d1.Year() == d2.Year() && d1.Month() == d2.Month() && d1.Day() == d2.Day()
}

func ComputeEaster(year int) time.Time {
    a := year % 19
    b := year / 100
    c := year % 100
    d := b / 4
    e := b % 4
    f := (b + 8) / 25
    g := (b - f + 1) / 3
    h := (19*a + b - d - g + 15) % 30
    i := c / 4
    k := c % 4
    l := (32 + 2*e + 2*i - h - k) % 7
    m := (a + 11*h + 22*l) / 451
    month := (h + l - 7*m + 114) / 31
    day := ((h + l - 7*m + 114) % 31) + 1
    return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
