package cmd

import (
	"ani-rem/models"
	"ani-rem/utils"
	"encoding/json"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var removeSpecificCalendarCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove calendar events for a specific anime",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := utils.NewGoogleCalendarClient()
		if err != nil {
			fmt.Printf("❌ Not connected: %v\n", err)
			fmt.Println("Run 'ani-rem calendar connect' first.")
			return
		}
		if !client.IsAuthenticated() {
			fmt.Println("❌ Not authenticated. Run 'ani-rem calendar connect'.")
			return
		}

		calendarID, err := client.GetPrimaryCalendarID()
		if err != nil {
			fmt.Printf("❌ Failed to get primary calendar: %v\n", err)
			return
		}

		// Get list of currently airing anime from local watchlist
		filePath := utils.GetStoragePath()
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println("❌ No saved anime found. Use 'ani-rem create' first.")
			return
		}
		var animes []models.AnimeData
		if err := json.Unmarshal(fileData, &animes); err != nil {
			fmt.Println("❌ Error reading anime list:", err)
			return
		}

		var airingTitles []string
		for _, a := range animes {
			if a.Status == "Currently Airing" {
				airingTitles = append(airingTitles, a.Title)
			}
		}
		if len(airingTitles) == 0 {
			fmt.Println("⚠️ No currently airing anime in your watchlist.")
			return
		}

		// Let user pick which anime to remove from calendar
		prompt := promptui.Select{
			Label: "Select anime to remove from Google Calendar",
			Items: airingTitles,
			Templates: &promptui.SelectTemplates{
				Active:   "➤ {{ . | cyan }}",
				Inactive: "  {{ . }}",
			},
		}
		idx, _, err := prompt.Run()
		if err != nil {
			fmt.Println("Cancelled.")
			return
		}
		selectedTitle := airingTitles[idx]

		// Confirm
		confirm := promptui.Prompt{
			Label:     fmt.Sprintf("Remove ALL calendar events for '%s'?", selectedTitle),
			IsConfirm: true,
		}
		if _, err := confirm.Run(); err != nil {
			fmt.Println("Removal cancelled.")
			return
		}

		fmt.Printf("🗑️  Removing events for %s...\n", selectedTitle)
		deleted, err := client.DeleteAnimeEvents(calendarID, selectedTitle)
		if err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
			return
		}
		if deleted == 0 {
			fmt.Printf("ℹ️  No events found for %s.\n", selectedTitle)
		} else {
			fmt.Printf("✅ Removed %d event(s) for %s.\n", deleted, selectedTitle)
		}
	},
}

func init() {
	calendarCmd.AddCommand(removeSpecificCalendarCmd)
}
