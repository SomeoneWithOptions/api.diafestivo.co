package holiday

import (
	"testing"
	"time"
)

func TestMakeHolidaysByYearKnownDates(t *testing.T) {
	tests := []struct {
		name string
		year int
		want map[string]string
	}{
		{
			name: "fixed monday shifted and easter derived holidays",
			year: 2025,
			want: map[string]string{
				"Año Nuevo":                     "2025-01-01",
				"el Día de los Reyes Magos":     "2025-01-06",
				"el Día de San José":            "2025-03-24",
				"Jueves Santo":                  "2025-04-17",
				"Viernes Santo":                 "2025-04-18",
				"la Ascensión del Señor":        "2025-06-02",
				"Corpus Christi":                "2025-06-23",
				"el Sagrado Corazón de Jesús":   "2025-06-30",
				"el Día de la Independencia":    "2025-07-20",
				"la Inmaculada Concepción":      "2025-12-08",
				"el Día de Navidad":             "2025-12-25",
				"la Independencia de Cartagena": "2025-11-17",
				"la Asunción de la Virgen":      "2025-08-18",
				"San Pedro y San Pablo":         "2025-06-30",
				"el Día de Nuestra Señora del Rosario de Chiquinquirá": "",
			},
		},
		{
			name: "chiquinquira holiday starts in 2026",
			year: 2026,
			want: map[string]string{
				"Jueves Santo":                "2026-04-02",
				"Viernes Santo":               "2026-04-03",
				"la Ascensión del Señor":      "2026-05-18",
				"Corpus Christi":              "2026-06-08",
				"el Sagrado Corazón de Jesús": "2026-06-15",
				"el Día de Nuestra Señora del Rosario de Chiquinquirá": "2026-07-13",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			holidays := MakeHolidaysByYear(tt.year)
			byName := map[string]string{}
			for _, holiday := range *holidays {
				byName[holiday.Name] = HolidayDateInCOT(holiday).Format(time.DateOnly)
			}

			for name, wantDate := range tt.want {
				gotDate, ok := byName[name]
				if wantDate == "" {
					if ok {
						t.Fatalf("%s should not exist for %d", name, tt.year)
					}
					continue
				}
				if !ok {
					t.Fatalf("missing holiday %q", name)
				}
				if gotDate != wantDate {
					t.Fatalf("%s date = %s, want %s", name, gotDate, wantDate)
				}
			}
		})
	}
}

func TestMakeHolidaysByYearReturnsClone(t *testing.T) {
	holidayList := MakeHolidaysByYear(2027)
	originalName := (*holidayList)[0].Name
	(*holidayList)[0].Name = "corrupted"

	freshHolidayList := MakeHolidaysByYear(2027)
	if got := (*freshHolidayList)[0].Name; got != originalName {
		t.Fatalf("cache was mutated: got %q, want %q", got, originalName)
	}
}

func TestClockDependentHolidayMethods(t *testing.T) {
	now := time.Date(2025, 6, 10, 12, 0, 0, 0, cotLocation)
	restore := SetNowFuncForTest(func() time.Time { return now })
	defer restore()

	corpusChristi := Holiday{Date: time.Date(2025, 6, 23, 0, 0, 0, 0, time.UTC), Name: "Corpus Christi"}
	if corpusChristi.IsToday() {
		t.Fatal("Corpus Christi should not be today on 2025-06-10")
	}
	if got, want := corpusChristi.DaysUntil(), 13; got != want {
		t.Fatalf("DaysUntil = %d, want %d", got, want)
	}

	holidays := MakeHolidaysByYear(2025)
	nextHoliday := holidays.FindNext()
	if nextHoliday == nil {
		t.Fatal("FindNext returned nil")
	}
	if nextHoliday.Name != "Corpus Christi" {
		t.Fatalf("FindNext = %q, want Corpus Christi", nextHoliday.Name)
	}

	remainingHolidays := holidays.GetRemaining()
	if len(*remainingHolidays) == 0 || (*remainingHolidays)[0].Name != "Corpus Christi" {
		t.Fatalf("first remaining holiday = %#v, want Corpus Christi", remainingHolidays)
	}
}

func TestGetRemainingExcludesToday(t *testing.T) {
	now := time.Date(2025, 6, 23, 12, 0, 0, 0, cotLocation)
	restore := SetNowFuncForTest(func() time.Time { return now })
	defer restore()

	holidays := MakeHolidaysByYear(2025)
	remainingHolidays := holidays.GetRemaining()
	if len(*remainingHolidays) == 0 {
		t.Fatal("expected remaining holidays")
	}
	if got := (*remainingHolidays)[0].Name; got == "Corpus Christi" {
		t.Fatalf("today's holiday should be excluded from /left semantics, got %q", got)
	}
}
