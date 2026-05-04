package utils

import "time"

// BroadcastDayMap maps Jikan broadcast day strings to time.Weekday
var BroadcastDayMap = map[string]time.Weekday{
	"Mondays":    time.Monday,
	"Tuesdays":   time.Tuesday,
	"Wednesdays": time.Wednesday,
	"Thursdays":  time.Thursday,
	"Fridays":    time.Friday,
	"Saturdays":  time.Saturday,
	"Sundays":    time.Sunday,
}
