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
	DaysUntil int32  `json:"daysUntil"`
}

type Holiday struct {
	Name string `json:"name"`
	Date string `json:"date"`
}

func NewNextHoliday(name string, date string, is_today bool, days_until int32) NextHoliday {
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
	current_time := time.Now()
	for _, h := range holidays {
		holiday_date, _ := time.Parse(time.RFC3339, h.Date)
		if holiday_date.After(current_time) {
			return &h
		}
	}
	return nil
}

func (h Holiday) IsToday() bool {
	holidayDate, _ := time.Parse(time.RFC3339, h.Date)

	currentDate := time.Now()

	return holidayDate.Year() == currentDate.Year() &&
		holidayDate.Month() == currentDate.Month() &&
		holidayDate.Day() == currentDate.Day()
}

func (h Holiday) DaysUntil() int {
	holidayDate, _ := time.Parse(time.RFC3339, h.Date)
	currentDate := time.Now()
	daysUntil := math.Ceil(holidayDate.Sub(currentDate).Hours() / 24)
	return int(daysUntil)
}
