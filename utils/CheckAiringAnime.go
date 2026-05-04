package utils

import (
	"ani-rem/models"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func CheckAiringAnime() {
	filePath := GetStoragePath()
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	var animes []models.AnimeData
	json.Unmarshal(fileData, &animes)

	for _, anime := range animes {
		if anime.Status == "Currently Airing" {
			fmt.Println("here inside loop curretnly airing")
			remaining := GetTimeUntilAiring(anime.Status, anime.Broadcast.Time, anime.Broadcast.Day)
			fmt.Println(remaining)

			// 1. Strip the prefix and remove spaces
			rawDuration := strings.TrimPrefix(remaining, "Next episode in ")
			cleanDuration := strings.ReplaceAll(rawDuration, " ", "")

			// 2. Parse into a time.Duration
			d, err := time.ParseDuration(cleanDuration)
			if err != nil {
				fmt.Println("Error parsing duration:", err)
				return
			}

			// 3. Check if it is less than 24 hours
			if d < 24*time.Hour {
				fmt.Println("Status: Episode drops in less than 24 hours!")
				SendNotification(anime.Title, cleanDuration)
			}

			// 4. Get the exact time (Current time + Duration)
			exactTime := time.Now().Add(d)

			fmt.Printf("Duration: %v\n", d)
			fmt.Printf("Exact Release Time: %s\n", exactTime.Format("2006-01-02 15:04:05"))

		}
	}
}
