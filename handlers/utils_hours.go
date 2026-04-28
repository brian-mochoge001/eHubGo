package handlers

import (
	"encoding/json"
	"time"
)

type BusinessHours struct {
	Mon    *TimeRange `json:"mon"`
	Tue    *TimeRange `json:"tue"`
	Wed    *TimeRange `json:"wed"`
	Thu    *TimeRange `json:"thu"`
	Fri    *TimeRange `json:"fri"`
	Sat    *TimeRange `json:"sat"`
	Sun    *TimeRange `json:"sun"`
	Is24_7 bool       `json:"is_24_7"`
}

type TimeRange struct {
	Open  string `json:"open"`
	Close string `json:"close"`
}

func IsBusinessOpen(hoursJSON []byte) (bool, error) {
	var hours BusinessHours
	if err := json.Unmarshal(hoursJSON, &hours); err != nil {
		return false, err
	}

	if hours.Is24_7 {
		return true, nil
	}

	now := time.Now()
	dayKey := map[time.Weekday]string{
		time.Monday: "mon", time.Tuesday: "tue", time.Wednesday: "wed",
		time.Thursday: "thu", time.Friday: "fri", time.Saturday: "sat", time.Sunday: "sun",
	}[now.Weekday()]

	var tr *TimeRange
	switch dayKey {
	case "mon": tr = hours.Mon
	case "tue": tr = hours.Tue
	case "wed": tr = hours.Wed
	case "thu": tr = hours.Thu
	case "fri": tr = hours.Fri
	case "sat": tr = hours.Sat
	case "sun": tr = hours.Sun
	}

	if tr == nil {
		return false, nil
	}

	open, _ := time.Parse("15:04", tr.Open)
	close, _ := time.Parse("15:04", tr.Close)
	currentTime, _ := time.Parse("15:04", now.Format("15:04"))

	return currentTime.After(open) && currentTime.Before(close), nil
}
