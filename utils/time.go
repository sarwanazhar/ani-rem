package utils

import (
	"fmt"
	"time"
)

func GetTimeUntilAiring(status string, broadcastTime string, broadcastDay string) string {
	if status == "Finished Airing" {
		return "Completed"
	}
	if status == "Not yet aired" {
		return "Upcoming (Release date TBD)"
	}
	if broadcastTime == "" || broadcastDay == "" || broadcastDay == "Unknown" {
		return "Airing schedule unavailable"
	}

	loc, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(loc)

	targetDay, ok := BroadcastDayMap[broadcastDay]
	if !ok {
		return "Invalid schedule"
	}

	var hour, min int
	fmt.Sscanf(broadcastTime, "%d:%d", &hour, &min)

	daysUntil := (int(targetDay) - int(now.Weekday()) + 7) % 7
	nextAiring := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, loc)
	nextAiring = nextAiring.AddDate(0, 0, daysUntil)
	if nextAiring.Before(now) {
		nextAiring = nextAiring.AddDate(0, 0, 7)
	}

	diff := time.Until(nextAiring)
	hours := int(diff.Hours())
	mins := int(diff.Minutes()) % 60
	return fmt.Sprintf("Next episode in %dh %dm", hours, mins)
}
