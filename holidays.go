package main

import (
	"math"
	"slices"
	"time"
)

var (
	cotLocation = time.FixedZone("UTC-5", -5*60*60)
	nowFunc     = time.Now
)

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

func (h Holidays) Sort() {
	slices.SortStableFunc(h, func(a, b Holiday) int { return a.Date.Compare(b.Date) })
}

func (h Holidays) FindNext() *Holiday {
	now := NowInCOT()
	for i := range h {
		holidayDate := HolidayDateInCOT(h[i])
		if IsSameDate(now, holidayDate) || holidayDate.After(now) {
			return &h[i]
		}
	}
	return nil
}

func (h Holidays) GetRemaining() Holidays {
	var remaining Holidays
	now := NowInCOT()
	for _, holiday := range h {
		if HolidayDateInCOT(holiday).After(now) {
			remaining = append(remaining, holiday)
		}
	}
	return remaining
}

func (h Holiday) IsToday() bool {
	return IsSameDate(NowInCOT(), HolidayDateInCOT(h))
}

func (h Holiday) DaysUntil() int {
	return int(math.Ceil(HolidayDateInCOT(h).Sub(NowInCOT()).Hours() / 24))
}

func SetNowFuncForTest(fn func() time.Time) func() {
	previous := nowFunc
	if fn == nil {
		nowFunc = time.Now
	} else {
		nowFunc = fn
	}
	return func() { nowFunc = previous }
}

func NowInCOT() time.Time {
	return nowFunc().In(cotLocation)
}

func HolidayDateInCOT(h Holiday) time.Time {
	return time.Date(h.Date.Year(), h.Date.Month(), h.Date.Day(), 0, 0, 0, 0, cotLocation)
}

func IsSameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func FindUpcomingHoliday() NextHoliday {
	now := NowInCOT()
	holidays := MakeHolidaysByYear(now.Year())
	next := holidays.FindNext()
	if next == nil {
		holidays = MakeHolidaysByYear(now.Year() + 1)
		next = holidays.FindNext()
	}
	return NextHoliday{Name: next.Name, Date: next.Date, IsToday: next.IsToday(), DaysUntil: next.DaysUntil()}
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

func MoveToMonday(date time.Time) time.Time {
	if date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, (8-int(date.Weekday()))%7)
	}
	return date
}

func MakeHolidaysByYear(year int) Holidays {
	easter := ComputeEaster(year)
	holidays := Holidays{
		{Date: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "Año Nuevo"},
		{Date: MoveToMonday(time.Date(year, 1, 6, 0, 0, 0, 0, time.UTC)), Name: "el Día de los Reyes Magos"},
		{Date: MoveToMonday(time.Date(year, 3, 19, 0, 0, 0, 0, time.UTC)), Name: "el Día de San José"},
		{Date: easter.AddDate(0, 0, -3), Name: "Jueves Santo"},
		{Date: easter.AddDate(0, 0, -2), Name: "Viernes Santo"},
		{Date: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "el Día del Trabajo"},
		{Date: MoveToMonday(easter.AddDate(0, 0, 39)), Name: "la Ascensión del Señor"},
		{Date: MoveToMonday(easter.AddDate(0, 0, 60)), Name: "Corpus Christi"},
		{Date: MoveToMonday(easter.AddDate(0, 0, 68)), Name: "el Sagrado Corazón de Jesús"},
		{Date: MoveToMonday(time.Date(year, 6, 29, 0, 0, 0, 0, time.UTC)), Name: "San Pedro y San Pablo"},
		{Date: time.Date(year, 7, 20, 0, 0, 0, 0, time.UTC), Name: "el Día de la Independencia"},
		{Date: time.Date(year, 8, 7, 0, 0, 0, 0, time.UTC), Name: "la Batalla de Boyacá"},
		{Date: MoveToMonday(time.Date(year, 8, 15, 0, 0, 0, 0, time.UTC)), Name: "la Asunción de la Virgen"},
		{Date: MoveToMonday(time.Date(year, 10, 12, 0, 0, 0, 0, time.UTC)), Name: "el Día de la Raza"},
		{Date: MoveToMonday(time.Date(year, 11, 1, 0, 0, 0, 0, time.UTC)), Name: "Todos los Santos"},
		{Date: MoveToMonday(time.Date(year, 11, 11, 0, 0, 0, 0, time.UTC)), Name: "la Independencia de Cartagena"},
		{Date: time.Date(year, 12, 8, 0, 0, 0, 0, time.UTC), Name: "la Inmaculada Concepción"},
		{Date: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "el Día de Navidad"},
	}
	if year >= 2026 {
		holidays = append(holidays, Holiday{
			Date: MoveToMonday(time.Date(year, 7, 9, 0, 0, 0, 0, time.UTC)),
			Name: "el Día de Nuestra Señora del Rosario de Chiquinquirá",
		})
	}
	holidays.Sort()
	return holidays
}
