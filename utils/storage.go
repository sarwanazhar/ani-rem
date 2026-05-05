package utils

import (
	"ani-rem/models"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func GetStoragePath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "ani-rem")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "list.json")
}

func SaveAnimeToList(anime models.AnimeData) error {
	var list []models.AnimeData
	filePath := GetStoragePath()

	// 1. Attempt to read existing file (if it exists)
	if _, err := os.Stat(filePath); err == nil {
		content, err := os.ReadFile(filePath)
		if err == nil {
			if err := json.Unmarshal(content, &list); err != nil {
				// Corrupted file – back it up and start fresh
				backupPath := filePath + ".bak"
				if copyErr := os.WriteFile(backupPath, content, 0644); copyErr == nil {
					fmt.Printf("⚠️  Corrupted list detected, backed up to %s\n", backupPath)
				}
				fmt.Printf("Starting with a fresh list due to: %v\n", err)
				list = nil
			}
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

func UpdateFullList(list []models.AnimeData) error {
	filePath := GetStoragePath()
	fileData, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, fileData, 0644)
}
