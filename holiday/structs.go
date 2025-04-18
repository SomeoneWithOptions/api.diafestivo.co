package holiday

import "time"

type NextHoliday struct {
	Name      string `json:"name"`
	Date      time.Time `json:"date"`
	IsToday   bool   `json:"isToday"`
	DaysUntil int    `json:"daysUntil"`
}

type Holiday struct {
	Date time.Time `json:"name"`
	Name string    `json:"date"`
}

type Holidays []Holiday

func NewNextHoliday(name string, date time.Time, is_today bool, days_until int) NextHoliday {
	var next_holiday NextHoliday
	next_holiday.Name = name
	next_holiday.Date = date
	next_holiday.IsToday = is_today
	next_holiday.DaysUntil = days_until
	return next_holiday
}

func NewHoliday(date time.Time, name string) Holiday {
	var h Holiday
	h.Date = date
	h.Name = name
	return h
}