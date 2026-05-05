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
		// No list yet is not an error
		return
	}

	var animes []models.AnimeData
	if err := json.Unmarshal(fileData, &animes); err != nil {
		fmt.Printf("Error parsing anime list: %v\n", err)
		return
	}

	// Load config to get the notification threshold (in hours)
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Printf("Could not load config, using default threshold (24h): %v\n", err)
		cfg = defaultConfig()
	}
	thresholdDuration := time.Duration(cfg.NotificationThresholdHours) * time.Hour

	for _, anime := range animes {
		if anime.Status != "Currently Airing" {
			continue
		}

		remaining := GetTimeUntilAiring(anime.Status, anime.Broadcast.Time, anime.Broadcast.Day)
		fmt.Println(remaining)

		// Strip the prefix and remove spaces
		rawDuration := strings.TrimPrefix(remaining, "Next episode in ")
		cleanDuration := strings.ReplaceAll(rawDuration, " ", "")

		d, err := time.ParseDuration(cleanDuration)
		if err != nil {
			fmt.Printf("Error parsing duration for %s: %v\n", anime.Title, err)
			continue
		}

		if d < thresholdDuration {
			fmt.Printf("Status: Episode drops in less than %d hours!\n", cfg.NotificationThresholdHours)
			SendNotification(anime.Title, cleanDuration)
		}

		exactTime := time.Now().Add(d)
		fmt.Printf("Duration: %v\n", d)
		fmt.Printf("Exact Release Time: %s\n", exactTime.Format("2006-01-02 15:04:05"))
	}
}
