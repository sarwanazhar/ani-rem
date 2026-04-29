package utils

import (
	"fmt"
	"time"
)

func GetTimeUntilAiring(status string, broadcastTime string, broadcastDay string) string {
	// 1. First check the status
	if status == "Finished Airing" {
		return "Completed"
	}
	if status == "Not yet aired" {
		return "Upcoming (Release date TBD)"
	}

	// 2. Validate broadcast data
	if broadcastTime == "" || broadcastDay == "" || broadcastDay == "Unknown" {
		return "Airing schedule unavailable"
	}

	// 3. Time calculation logic (JST)
	loc, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(loc)

	dayMap := map[string]time.Weekday{
		"Mondays":    time.Monday,
		"Tuesdays":   time.Tuesday,
		"Wednesdays": time.Wednesday,
		"Thursdays":  time.Thursday,
		"Fridays":    time.Friday,
		"Saturdays":  time.Saturday,
		"Sundays":    time.Sunday,
	}

	targetDay, ok := dayMap[broadcastDay]
	if !ok {
		return "Invalid schedule"
	}

	var hour, min int
	fmt.Sscanf(broadcastTime, "%d:%d", &hour, &min)

	// Calculate next episode time
	daysUntil := (int(targetDay) - int(now.Weekday()) + 7) % 7
	nextAiring := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, loc)
	nextAiring = nextAiring.AddDate(0, 0, daysUntil)

	// If it aired today but the time passed, it's next week
	if nextAiring.Before(now) {
		nextAiring = nextAiring.AddDate(0, 0, 7)
	}

	diff := time.Until(nextAiring)

	// 4. Return the formatted countdown
	hours := int(diff.Hours())
	mins := int(diff.Minutes()) % 60
	return fmt.Sprintf("Next episode in %dh %dm", hours, mins)
}
