package templateinfo

import "time"

type TemplateInfo struct {
	IsToday   bool
	Name      string
	DaysUntil int
	Date      time.Time
	GifURL    *string
	Day       int
	Month     string
	Year      int
	Weekday   string
}

func NewTemplateInfo(name string, isToday bool, daysUntil int, date time.Time, gifURL *string, day int, month string, year int, weekDay string) TemplateInfo {
	return TemplateInfo{
		Name:      name,
		IsToday:   isToday,
		DaysUntil: daysUntil,
		Date:      date,
		GifURL:    gifURL,
		Day:       day,
		Month:     month,
		Year:      year,
		Weekday:   weekDay,
	}
}
