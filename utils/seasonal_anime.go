package utils

import (
	"ani-rem/models"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	SeasonWinter = "winter"
	SeasonSpring = "spring"
	SeasonSummer = "summer"
	SeasonFall   = "fall"
)

// FetchSeasonalAnime fetches anime from a specific season/year or "now" for current season
func FetchSeasonalAnime(year int, season string, page int) (*models.SeasonalResponse, error) {
	var apiURL string

	if year == 0 && season == "now" {
		// Fetch currently airing season
		apiURL = "https://api.jikan.moe/v4/seasons/now"
	} else {
		apiURL = fmt.Sprintf("https://api.jikan.moe/v4/seasons/%d/%s?page=%d", year, season, page)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jikan API returned %d: %s", resp.StatusCode, resp.Status)
	}

	var result models.SeasonalResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &result, nil
}

// FetchCurrentlyAiringSeason fetches anime from the current season only
func FetchCurrentlyAiringSeason() ([]models.AnimeData, error) {
	resp, err := FetchSeasonalAnime(0, "now", 1)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// FilterAiringOnly filters seasonal anime to only include "Currently Airing" status
func FilterAiringOnly(animes []models.AnimeData) []models.AnimeData {
	var airing []models.AnimeData
	for _, a := range animes {
		if a.Status == "Currently Airing" {
			airing = append(airing, a)
		}
	}
	return airing
}

// GetSeasonName returns the current season name based on month
func GetSeasonName() (string, int) {
	month := time.Now().Month()
	year := time.Now().Year()

	switch month {
	case 1, 2, 3:
		return SeasonWinter, year
	case 4, 5, 6:
		return SeasonSpring, year
	case 7, 8, 9:
		return SeasonSummer, year
	case 10, 11, 12:
		return SeasonFall, year
	default:
		return SeasonWinter, year
	}
}

// BulkAddAnimeToList adds multiple anime to the watchlist, skipping duplicates
func BulkAddAnimeToList(animes []models.AnimeData) (added, skipped int, err error) {
	filePath := GetStoragePath()
	var list []models.AnimeData

	// Read existing list
	if _, err := os.Stat(filePath); err == nil {
		content, err := os.ReadFile(filePath)
		if err == nil {
			if err := json.Unmarshal(content, &list); err != nil {
				backupPath := filePath + ".bak"
				if copyErr := os.WriteFile(backupPath, content, 0644); copyErr == nil {
					fmt.Printf("⚠️  Corrupted list detected, backed up to %s\n", backupPath)
				}
				list = nil
			}
		}
	}

	// Track existing MAL IDs for deduplication
	existingIDs := make(map[int]bool)
	for _, item := range list {
		existingIDs[item.MalID] = true
	}

	// Add new anime
	for _, anime := range animes {
		if existingIDs[anime.MalID] {
			skipped++
			continue
		}
		list = append(list, anime)
		existingIDs[anime.MalID] = true
		added++
	}

	// Save updated list
	fileData, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return 0, 0, err
	}

	if err := os.WriteFile(filePath, fileData, 0644); err != nil {
		return 0, 0, err
	}

	return added, skipped, nil
}
