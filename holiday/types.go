package holiday

import "time"

type NextHoliday struct {
	Name      string    `json:"name"`
	Date      time.Time `json:"date"`
	IsToday   bool      `json:"isToday"`
	DaysUntil int       `json:"daysUntil"`
}

type Holiday struct {
	Date time.Time `json:"date"`
	Name string    `json:"name"`
}

type Holidays []Holiday

func NewNextHoliday(name string, date time.Time, isToday bool, daysUntil int) NextHoliday {
	return NextHoliday{Name: name, Date: date, IsToday: isToday, DaysUntil: daysUntil}
}

func NewHoliday(date time.Time, name string) Holiday {
	return Holiday{Date: date, Name: name}
}
