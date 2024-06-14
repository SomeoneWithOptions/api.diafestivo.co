package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/SomeoneWithOptions/api.diafestivo.co/holiday"
)

func TestIsToday(t *testing.T) {
	today := time.Now()

	h := holiday.Holiday{
		Name: "Test Holiday",
		Date: today.Format("2006-01-02T15:04:05Z07:00"),
	}


	fmt.Println(today)
	fmt.Println(h.Date)

	if !h.IsToday() {
		t.Errorf("today %v holiday date %v", today, h.Date)
	}

	future := time.Date(today.Year(), today.Month(), today.Day()+1, 0, 0, 0, 0, time.UTC)
	h = holiday.Holiday{
		Name: "Test Holiday",
		Date: future.Format("2006-01-02"),
	}
	if h.IsToday() {
		t.Errorf("IsToday() = true; want false for a future date")
	}

	past := time.Date(today.Year(), today.Month(), today.Day()-1, 0, 0, 0, 0, time.UTC)
	h = holiday.Holiday{
		Name: "Test Holiday",
		Date: past.Format("2006-01-02"),
	}
	if h.IsToday() {
		t.Errorf("IsToday() = true; want false for a past date")
	}
}
