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

var (
	syncWeeks    int
	syncCalendar string
	syncAll      bool
	syncAnime    string
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync anime airing schedule to Google Calendar",
	Long: `Sync your currently airing anime to Google Calendar.

Examples:
  ani-rem sync                    # Interactive mode
  ani-rem sync --all              # Sync all currently airing anime
  ani-rem sync --anime "One Piece" # Sync specific anime
  ani-rem sync --weeks 24         # Sync next 24 weeks`,
	Run: func(cmd *cobra.Command, args []string) {
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
		if len(animes) == 0 {
			fmt.Println("📭 Your watchlist is empty.")
			return
		}

		var airingAnimes []models.AnimeData
		for _, a := range animes {
			if a.Status == "Currently Airing" {
				airingAnimes = append(airingAnimes, a)
			}
		}
		if len(airingAnimes) == 0 {
			fmt.Println("⚠️ No currently airing anime found. Only those can be synced.")
			return
		}

		var selected []models.AnimeData
		if syncAll {
			selected = airingAnimes
		} else if syncAnime != "" {
			found := false
			for _, a := range airingAnimes {
				if a.Title == syncAnime {
					selected = append(selected, a)
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("❌ Anime '%s' not found or not currently airing.\n", syncAnime)
				return
			}
		} else {
			items := append([]models.AnimeData{{Title: "📅 SYNC ALL CURRENTLY AIRING ANIME"}}, airingAnimes...)
			prompt := promptui.Select{
				Label: "Select anime to sync",
				Items: items,
				Templates: &promptui.SelectTemplates{
					Active:   "➤ {{ .Title | cyan }}",
					Inactive: "  {{ .Title }}",
				},
			}
			idx, _, err := prompt.Run()
			if err != nil {
				return
			}
			if idx == 0 {
				selected = airingAnimes
			} else {
				selected = []models.AnimeData{airingAnimes[idx-1]}
			}
		}

		client, err := utils.NewGoogleCalendarClient()
		if err != nil {
			fmt.Printf("⚠️ Calendar not set up: %v\n", err)
			fmt.Println("\nRun 'ani-rem calendar connect' to sign in to Google Calendar first.")
			return
		}
		if !client.IsAuthenticated() {
			fmt.Println("❌ Not connected to Google Calendar.")
			fmt.Println("Run 'ani-rem calendar connect' to sign in.")
			return
		}

		calendarID := syncCalendar
		if calendarID == "" {
			id, err := client.GetPrimaryCalendarID()
			if err != nil {
				fmt.Printf("❌ Failed to get calendar: %v\n", err)
				return
			}
			calendarID = id
			fmt.Println("✓ Using primary calendar")
		}

		confirm := promptui.Prompt{
			Label:     fmt.Sprintf("Sync %d anime(s) for %d weeks to Google Calendar", len(selected), syncWeeks),
			IsConfirm: true,
		}
		if _, err := confirm.Run(); err != nil {
			fmt.Println("Sync cancelled.")
			return
		}

		fmt.Println("\n🔄 Syncing to Google Calendar...")
		err = client.SyncMultipleAnime(selected, syncWeeks, calendarID)
		if err != nil {
			fmt.Printf("❌ Sync failed: %v\n", err)
			return
		}
		fmt.Println("\n✅ Sync completed! Open Google Calendar to see your schedule.")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().IntVarP(&syncWeeks, "weeks", "w", 12, "Number of weeks to schedule")
	syncCmd.Flags().StringVarP(&syncCalendar, "calendar", "c", "", "Calendar ID (default: primary)")
	syncCmd.Flags().BoolVarP(&syncAll, "all", "a", false, "Sync all currently airing anime")
	syncCmd.Flags().StringVarP(&syncAnime, "anime", "n", "", "Sync specific anime by title")
}
