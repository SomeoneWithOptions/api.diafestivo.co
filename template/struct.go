package template

type TemplateInfo struct {
	IsToday   bool
	Name      string
	DaysUntil int32
}

func NewTemplateInfo(name string, is_today bool, days_until int32) TemplateInfo {
	var template_info TemplateInfo
	template_info.Name = name
	template_info.IsToday = is_today
	template_info.DaysUntil = days_until
	return template_info
}
