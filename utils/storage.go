package utils

import (
	"ani-rem/models"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// This helper ensures the directory exists and returns the full path
func GetStoragePath() string {
	// Option A: Use a hidden folder in Home (Best for persistence)
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "ani-rem")

	// Option B: Use /tmp (Files deleted on reboot)
	// dir := "/tmp/ani-rem"

	// Create the directory if it doesn't exist (mkdir -p)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Println("Error creating config directory:", err)
	}

	return filepath.Join(dir, "list.json")
}

func SaveAnimeToList(anime models.AnimeData) error {
	var list []models.AnimeData
	filePath := GetStoragePath()

	// 1. Read existing file
	if _, err := os.Stat(filePath); err == nil {
		content, err := os.ReadFile(filePath)
		if err == nil {
			json.Unmarshal(content, &list)
		}
	}

	// 2. Prevent duplicates
	for _, item := range list {
		if item.MalID == anime.MalID {
			fmt.Printf("⚠️  %s is already in your list!\n", anime.Title)
			return nil
		}
	}

	// 3. Append and Write
	list = append(list, anime)
	fileData, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("💾 Saving to: %s\n", filePath)
	return os.WriteFile(filePath, fileData, 0644)
}

// this function is to handle overwriting after deletion
func UpdateFullList(list []models.AnimeData) error {
	filePath := GetStoragePath()
	fileData, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, fileData, 0644)
}
