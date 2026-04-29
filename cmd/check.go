package cmd

import (
	"ani-rem/models"
	"ani-rem/utils"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for airing anime and send notifications",
	Run: func(cmd *cobra.Command, args []string) {
		filePath := utils.GetStoragePath()
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			// If file doesn't exist, just exit quietly
			return
		}

		var animes []models.AnimeData
		json.Unmarshal(fileData, &animes)

		fmt.Println("🚀 Running airing check...")

		for _, anime := range animes {
			// We only care about countdowns for shows that are currently airing
			if anime.Status == "Currently Airing" {
				// Update: Passing anime.Status as the first argument
				remaining := utils.GetTimeUntilAiring(anime.Status, anime.Broadcast.Time, anime.Broadcast.Day)

				fmt.Printf("🔍 %s: %s\n", anime.Title, remaining)

				// Logic: If the string contains "0h", it means the episode airs in 0-59 minutes.
				// We also check for "Next episode in" to make sure we aren't matching an error string.
				if strings.Contains(remaining, "Next episode in 0h") {
					utils.SendNotification("Anime Airing Soon!", fmt.Sprintf("%s airs in %s", anime.Title, strings.TrimPrefix(remaining, "Next episode in ")))
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
