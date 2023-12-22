package templateinfo

type TemplateInfo struct {
	IsToday   bool
	Name      string
	DaysUntil int32
	Date      string
	GifURL    string
}

func NewTemplateInfo(name string, is_today bool, days_until int32, date string, gif_url string) TemplateInfo {
	var template_info TemplateInfo
	template_info.Name = name
	template_info.IsToday = is_today
	template_info.DaysUntil = days_until
	template_info.Date = date
	template_info.GifURL = gif_url
	return template_info
}
