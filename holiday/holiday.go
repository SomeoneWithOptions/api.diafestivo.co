package holiday

import "fmt"

type NextHoliday struct {
	Holiday   Holiday
	IsToday   bool
	DaysUntil int32
}

type Holiday struct {
	Name string `json:"name"`
	Date string `json:"date"`
}

func NewNextHoliday(name string, date string, is_today bool, days_until int32) NextHoliday {
	var next_holiday NextHoliday
	next_holiday.Holiday.Name = name
	next_holiday.Holiday.Date = date
	next_holiday.IsToday = is_today
	next_holiday.DaysUntil = days_until
	return next_holiday
}

func (n NextHoliday) Print() string {
	return fmt.Sprintf("name: %s\ndate: %s\nisToday: %v\ndaysUntil: %d", n.Holiday.Name, n.Holiday.Date, n.IsToday, n.DaysUntil)
}
