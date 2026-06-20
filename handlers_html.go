package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/SomeoneWithOptions/api.diafestivo.co/giphy"
	"github.com/SomeoneWithOptions/api.diafestivo.co/holiday"
	"github.com/SomeoneWithOptions/api.diafestivo.co/templateinfo"
)

func handleTemplate(w http.ResponseWriter, r *http.Request) {
	var gifURL *string
	nextHoliday := holiday.FindUpcomingHoliday()

	if nextHoliday.IsToday {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		var err error
		gifURL, err = giphy.FetchGifURLContext(ctx)
		if err != nil {
			slog.Error("failed to fetch giphy gif", "error", err)
		}
	}

	holidayDate := holiday.HolidayDateInCOT(holiday.Holiday{Date: nextHoliday.Date})
	templateInfo := templateinfo.NewTemplateInfo(
		nextHoliday.Name,
		nextHoliday.IsToday,
		nextHoliday.DaysUntil,
		nextHoliday.Date,
		gifURL,
		holidayDate.Day(),
		monthNames[int(holidayDate.Month())],
		holidayDate.Year(),
		weekdayNames[int(holidayDate.Weekday())],
	)

	renderTemplate(w, indexTemplate, templateInfo)
}

func handleLeft(w http.ResponseWriter, _ *http.Request) {
	type leftHolidayView struct {
		Name     string
		Day      int
		DaysLeft int
		WeekDay  string
		Month    string
	}

	now := holiday.NowInCOT()
	year := now.Year()
	holidays := holiday.MakeHolidaysByYear(year)
	remainingHolidays := holidays.GetRemaining()

	const minDaysToShow = 3
	if len(*remainingHolidays) < minDaysToShow {
		nextYearHolidays := holiday.MakeHolidaysByYear(year + 1)
		neededHolidays := minDaysToShow - len(*remainingHolidays)
		for i, holidayItem := range *nextYearHolidays {
			if i == neededHolidays {
				break
			}
			*remainingHolidays = append(*remainingHolidays, holidayItem)
		}
	}

	templateData := make([]leftHolidayView, 0, len(*remainingHolidays))
	for _, holidayItem := range *remainingHolidays {
		holidayDate := holiday.HolidayDateInCOT(holidayItem)
		templateData = append(templateData, leftHolidayView{
			Name:     holidayItem.Name,
			Day:      holidayDate.Day(),
			DaysLeft: holidayItem.DaysUntil(),
			WeekDay:  weekdayNames[int(holidayDate.Weekday())],
			Month:    monthNames[int(holidayDate.Month())],
		})
	}

	renderTemplate(w, leftTemplate, templateData)
}
