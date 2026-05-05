package cmd

import (
	"ani-rem/models"
	"ani-rem/utils"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	seasonalYear   int
	seasonalSeason string
	seasonalFilter bool
	seasonalPage   int
)

var seasonalCmd = &cobra.Command{
	Use:   "seasonal",
	Short: "Browse and bulk-add seasonal anime",
	Long: `Browse anime from the current or specified season.
Allows interactive selection and bulk adding to your watchlist.

Examples:
  ani-rem seasonal                    # Browse current season
  ani-rem seasonal --year 2024 --season spring
  ani-rem seasonal --filter           # Show only currently airing
  ani-rem seasonal --bulk             # Enable multi-select mode`,
	Run: func(cmd *cobra.Command, args []string) {
		runSeasonalBrowser()
	},
}

func runSeasonalBrowser() {
	fmt.Println("🔍 Fetching seasonal anime...")

	var animes []models.AnimeData
	var err error

	if seasonalYear == 0 && seasonalSeason == "" {
		animes, err = utils.FetchCurrentlyAiringSeason()
		seasonName, year := utils.GetSeasonName()
		fmt.Printf("📅 Showing %s %d season\n", strings.Title(seasonName), year)
	} else {
		if seasonalSeason == "" {
			seasonalSeason, seasonalYear = utils.GetSeasonName()
		}
		resp, err := utils.FetchSeasonalAnime(seasonalYear, seasonalSeason, seasonalPage)
		if err != nil {
			fmt.Printf("❌ Failed to fetch seasonal anime: %v\n", err)
			return
		}
		animes = resp.Data
		fmt.Printf("📅 Showing %s %d (Page %d)\n", strings.Title(seasonalSeason), seasonalYear, seasonalPage)
	}

	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	if len(animes) == 0 {
		fmt.Println("⚠️  No anime found for this season.")
		return
	}

	if seasonalFilter {
		animes = utils.FilterAiringOnly(animes)
		fmt.Printf("🎬 Filtered to %d currently airing anime\n", len(animes))
	}

	if len(animes) == 0 {
		fmt.Println("⚠️  No currently airing anime found after filtering.")
		return
	}

	// Prepare items for interactive selection
	items := make([]models.SeasonListItem, len(animes))
	for i, a := range animes {
		items[i] = models.SeasonListItem{
			AnimeData: a,
			Selected:  false,
		}
	}
	items = append(items, models.SeasonListItem{
		AnimeData: models.AnimeData{Title: "✅ Add Selected to Watchlist"},
	})
	items = append(items, models.SeasonListItem{
		AnimeData: models.AnimeData{Title: "🔄 Select All / Deselect All"},
	})
	items = append(items, models.SeasonListItem{
		AnimeData: models.AnimeData{Title: "↩️  Exit to Menu"},
	})

	for {
		prompt := promptui.Select{
			Label:    "Seasonal Anime (navigate with arrows, Enter to select)",
			Items:    items,
			HideHelp: true,
			Templates: &promptui.SelectTemplates{
				Active:   `{{ if .Selected }}☑{{ else }}☐{{ end }} {{ .Title | cyan }}{{ if .Score }} ({{ .Score | yellow }}){{ end }}{{ if .Status }} [{{ .Status | green }}]{{ end }}`,
				Inactive: `{{ if .Selected }}☑{{ else }}☐{{ end }} {{ .Title }}{{ if .Score }} ({{ .Score }}){{ end }}{{ if .Status }} [{{ .Status }}]{{ end }}`,
				Selected: `✔ {{ .Title | green }}`,
			},
		}

		index, _, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				os.Exit(0)
			}
			return
		}
		// ✨ Clear the prompt label line to prevent duplication
		fmt.Print("\033[1A\033[2K\r") // Move up 1 line, clear it, return to start

		// Handle special actions
		if index == len(animes) {
			// Add Selected
			var selected []models.AnimeData
			for _, item := range items[:len(animes)] {
				if item.Selected {
					selected = append(selected, item.AnimeData)
				}
			}
			if len(selected) == 0 {
				fmt.Println("⚠️  No anime selected. Use the detail view to select items.")
				continue
			}
			confirm := promptui.Prompt{
				Label:     fmt.Sprintf("Add %d selected anime to watchlist?", len(selected)),
				IsConfirm: true,
			}
			if _, err := confirm.Run(); err != nil {
				fmt.Println("Cancelled.")
				continue
			}
			added, skipped, err := utils.BulkAddAnimeToList(selected)
			if err != nil {
				fmt.Printf("❌ Error adding anime: %v\n", err)
			} else {
				fmt.Printf("✅ Added %d anime, skipped %d (already in list)\n", added, skipped)
			}
			continue
		}
		if index == len(animes)+1 {
			// Toggle All
			allSelected := true
			for _, item := range items[:len(animes)] {
				if !item.Selected {
					allSelected = false
					break
				}
			}
			for i := range items {
				if i < len(animes) {
					items[i].Selected = !allSelected
				}
			}
			action := map[bool]string{true: "Deselected", false: "Selected"}[allSelected]
			fmt.Printf("🔄 %s all anime\n", action)
			continue
		}
		if index == len(animes)+2 {
			return
		}

		// Handle individual anime - show detail submenu
		selectedItem := &items[index]
		runAnimeDetailMenu(selectedItem)
	}
}

// runAnimeDetailMenu shows details and allows selection of a single anime
func runAnimeDetailMenu(item *models.SeasonListItem) {
	for {
		checkbox := map[bool]string{true: "☑", false: "☐"}[item.Selected]
		actionPrompt := promptui.Select{
			Label: fmt.Sprintf("%s - Actions", item.Title),
			Items: []string{
				fmt.Sprintf("%s Toggle Selection", checkbox),
				"📖 Show Synopsis",
				"➕ Add to Watchlist",
				"⬅️  Back to List",
			},
			HideHelp: true,
		}
		_, action, err := actionPrompt.Run()
		if err != nil {
			return
		}

		switch action {
		case "☑ Toggle Selection", "☐ Toggle Selection":
			item.Selected = !item.Selected
			newCheckbox := map[bool]string{true: "☑", false: "☐"}[item.Selected]
			status := map[bool]string{true: "Selected ✓", false: "Deselected ✗"}[item.Selected]
			fmt.Printf("  %s %s: %s\n", newCheckbox, item.Title, status)
			return
		case "📖 Show Synopsis":
			fmt.Printf("\n--- %s ---\n", item.Title)
			fmt.Printf("Status: %s | Score: %.2f\n", item.Status, item.Score)
			fmt.Println("\nSynopsis:", item.Synopsis)
			fmt.Print("\nPress Enter to continue...")
			fmt.Scanln()
		case "➕ Add to Watchlist":
			err := utils.SaveAnimeToList(item.AnimeData)
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			} else {
				fmt.Println("✅ Added to watchlist!")
			}
			return
		case "⬅️  Back to List":
			return
		}
	}
}

func init() {
	rootCmd.AddCommand(seasonalCmd)
	seasonalCmd.Flags().IntVarP(&seasonalYear, "year", "y", 0, "Year for seasonal anime (default: current)")
	seasonalCmd.Flags().StringVarP(&seasonalSeason, "season", "s", "", "Season: winter, spring, summer, fall (default: current)")
	seasonalCmd.Flags().BoolVarP(&seasonalFilter, "filter", "f", false, "Show only currently airing anime")
	seasonalCmd.Flags().IntVarP(&seasonalPage, "page", "p", 1, "Page number for pagination")
}
