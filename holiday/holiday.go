package holiday

import (
	"math"
	"slices"
	"sync"
	"time"
)

// Package-level timezone avoids recreating it on every call.
var cotLocation = time.FixedZone("UTC-5", -5*60*60)

// nowFunc is a test seam. Production uses time.Now.
var nowFunc = time.Now

// Cache computed holidays per year. Holidays are deterministic
// for a given year, so this never needs invalidation.
var holidayCache sync.Map

func (h Holidays) Sort() {
	slices.SortStableFunc(h, func(a, b Holiday) int {
		return a.Date.Compare(b.Date)
	})
}

func (h *Holidays) FindNext() *Holiday {
	now := NowInCOT()
	for _, holiday := range *h {
		holidayDate := HolidayDateInCOT(holiday)
		if IsSameDate(now, holidayDate) {
			return &holiday
		}
		if holidayDate.After(now) {
			return &holiday
		}
	}
	return nil
}

func (h *Holidays) GetRemaining() *Holidays {
	remainingHolidays := Holidays{}
	now := NowInCOT()

	for _, holiday := range *h {
		holidayDate := HolidayDateInCOT(holiday)
		if holidayDate.After(now) {
			remainingHolidays = append(remainingHolidays, holiday)
		}
	}
	return &remainingHolidays
}

func (h Holiday) IsToday() bool {
	now := NowInCOT()
	holidayDate := HolidayDateInCOT(h)
	return IsSameDate(now, holidayDate)
}

func (h Holiday) DaysUntil() int {
	now := NowInCOT()
	holidayDate := HolidayDateInCOT(h)
	daysUntil := math.Ceil(holidayDate.Sub(now).Hours() / 24)
	return int(daysUntil)
}

// SetNowFuncForTest replaces the package clock and returns a restore function.
// It is intended for tests.
func SetNowFuncForTest(fn func() time.Time) func() {
	previousNowFunc := nowFunc
	if fn == nil {
		nowFunc = time.Now
	} else {
		nowFunc = fn
	}
	return func() { nowFunc = previousNowFunc }
}

// NowInCOT returns the current time in Colombia timezone.
func NowInCOT() time.Time {
	return nowFunc().In(cotLocation)
}

// HolidayDateInCOT returns a holiday's civil date at midnight in Colombia timezone.
func HolidayDateInCOT(h Holiday) time.Time {
	return time.Date(h.Date.Year(), h.Date.Month(), h.Date.Day(), 0, 0, 0, 0, cotLocation)
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
	holidays := MakeHolidaysByYear(now.Year())
	nextHoliday := holidays.FindNext()

	if nextHoliday == nil {
		holidays = MakeHolidaysByYear(now.Year() + 1)
		nextHoliday = holidays.FindNext()
	}

	result := NewNextHoliday(
		nextHoliday.Name,
		nextHoliday.Date,
		nextHoliday.IsToday(),
		nextHoliday.DaysUntil(),
	)
	return &result
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
	if cached, ok := holidayCache.Load(year); ok {
		cachedHolidays := cached.(Holidays)
		clonedHolidays := slices.Clone(cachedHolidays)
		return &clonedHolidays
	}

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
	canonicalHolidays := slices.Clone(holidays)
	holidayCache.Store(year, canonicalHolidays)
	return &holidays
}
