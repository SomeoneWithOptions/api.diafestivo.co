package templateinfo

type TemplateInfo struct {
	IsToday   bool
	Name      string
	DaysUntil int8
	Date      string
	GifURL    *string
	Day       int
	Month     string
	Year      int
	Weekday   string
}

func NewTemplateInfo(name string, is_today bool, days_until int8, date string, gif_url *string, day int, month string, year int, week_day string) TemplateInfo {
	var template_info TemplateInfo
	template_info.Name = name
	template_info.IsToday = is_today
	template_info.DaysUntil = days_until
	template_info.Date = date
	template_info.GifURL = gif_url
	template_info.Day = day
	template_info.Month = month
	template_info.Year = year
	template_info.Weekday = week_day
	return template_info
}
