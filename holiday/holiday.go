package holiday

import (
	"math"
	"sort"
	"sync"
	"time"
)

// Package-level timezone avoids recreating it on every call.
var cotLocation = time.FixedZone("UTC-5", -5*60*60)

// Cache computed holidays per year. Holidays are deterministic
// for a given year, so this never needs invalidation.
var holidayCache sync.Map

func (h Holidays) Sort() {
	sort.SliceStable(h, func(i, j int) bool {
		return h[i].Date.Before(h[j].Date)
	})
}

func (h *Holidays) FindNext() *Holiday {
	now := NowInCOT()
	for _, holiday := range *h {
		hDate := HolidayDateInCOT(holiday)
		if IsSameDate(now, hDate) {
			return &holiday
		}
		if hDate.After(now) {
			return &holiday
		}
	}
	return nil
}

func (h *Holidays) GetRemaining() *Holidays {
	var remainingHolidays Holidays
	now := NowInCOT()

	for _, holiday := range *h {
		hDate := HolidayDateInCOT(holiday)
		if hDate.After(now) {
			remainingHolidays = append(remainingHolidays, holiday)
		}
	}
	return &remainingHolidays
}

func (h Holiday) IsToday() bool {
	now := NowInCOT()
	hDate := HolidayDateInCOT(h)
	return IsSameDate(now, hDate)
}

func (h Holiday) DaysUntil() int {
	now := NowInCOT()
	hDate := HolidayDateInCOT(h)
	daysUntil := math.Ceil(hDate.Sub(now).Hours() / 24)
	return int(daysUntil)
}

// NowInCOT returns the current time in Colombia timezone.
func NowInCOT() time.Time {
	return time.Now().In(cotLocation)
}

// HolidayDateInCOT converts a holiday's UTC date to Colombia timezone.
func HolidayDateInCOT(h Holiday) time.Time {
	return h.Date.In(cotLocation).Add(time.Hour * 5)
}

// MakeDatesInCOT returns (currentTime, holidayDate) in COT.
// Kept for backward compatibility.
func MakeDatesInCOT(h Holiday) (time.Time, time.Time) {
	return NowInCOT(), HolidayDateInCOT(h)
}

func IsSameDate(d1, d2 time.Time) bool {
	return d1.Year() == d2.Year() && d1.Month() == d2.Month() && d1.Day() == d2.Day()
}

func FindUpcomingHoliday() *NextHoliday {
	now := NowInCOT()
	allHolidays := MakeHolidaysByYear(now.Year())
	nextHoliday := allHolidays.FindNext()

	if nextHoliday == nil {
		allHolidays = MakeHolidaysByYear(now.Year() + 1)
		nextHoliday = allHolidays.FindNext()
	}

	n := NewNextHoliday(
		nextHoliday.Name,
		nextHoliday.Date,
		nextHoliday.IsToday(),
		nextHoliday.DaysUntil(),
	)
	return &n
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

func MoveToMonday(t time.Time) time.Time {
	if t.Weekday() != time.Monday {
		days := (8 - int(t.Weekday())) % 7
		t = t.AddDate(0, 0, days)
	}
	return t
}

func MakeHolidaysByYear(year int) *Holidays {
	// Check cache first.
	if cached, ok := holidayCache.Load(year); ok {
		return cached.(*Holidays)
	}

	e := ComputeEaster(year)
	h := Holidays{
		{time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), "Año Nuevo"},
		{MoveToMonday(time.Date(year, 1, 6, 0, 0, 0, 0, time.UTC)), "el Día de los Reyes Magos"},
		{MoveToMonday(time.Date(year, 3, 19, 0, 0, 0, 0, time.UTC)), "el Día de San José"},
		{e.AddDate(0, 0, -3), "Jueves Santo"},
		{e.AddDate(0, 0, -2), "Viernes Santo"},
		{time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), "el Día del Trabajo"},
		{MoveToMonday(e.AddDate(0, 0, 39)), "la Ascensión del Señor"},
		{MoveToMonday(e.AddDate(0, 0, 60)), "Corpus Christi"},
		{MoveToMonday(e.AddDate(0, 0, 68)), "el Sagrado Corazón de Jesús"},
		{MoveToMonday(time.Date(year, 6, 29, 0, 0, 0, 0, time.UTC)), "San Pedro y San Pablo"},
		{time.Date(year, 7, 20, 0, 0, 0, 0, time.UTC), "el Día de la Independencia"},
		{time.Date(year, 8, 7, 0, 0, 0, 0, time.UTC), "la Batalla de Boyacá"},
		{MoveToMonday(time.Date(year, 8, 15, 0, 0, 0, 0, time.UTC)), "la Asunción de la Virgen"},
		{MoveToMonday(time.Date(year, 10, 12, 0, 0, 0, 0, time.UTC)), "el Día de la Raza"},
		{MoveToMonday(time.Date(year, 11, 1, 0, 0, 0, 0, time.UTC)), "Todos los Santos"},
		{MoveToMonday(time.Date(year, 11, 11, 0, 0, 0, 0, time.UTC)), "la Independencia de Cartagena"},
		{time.Date(year, 12, 8, 0, 0, 0, 0, time.UTC), "la Inmaculada Concepción"},
		{time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), "el Día de Navidad"},
	}

	h.Sort()

	holidayCache.Store(year, &h)
	return &h
}
