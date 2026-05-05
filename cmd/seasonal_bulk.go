package cmd

import (
	"ani-rem/models"
	"ani-rem/utils"
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var seasonalBulkCmd = &cobra.Command{
	Use:   "bulk",
	Short: "Quick bulk-add currently airing seasonal anime",
	Long: `Fetch the current season's anime and allow quick bulk selection.
Ideal for adding multiple new season shows at once.

Examples:
  ani-rem seasonal bulk                    # Interactive multi-select
  ani-rem seasonal bulk --all              # Add all currently airing
  ani-rem seasonal bulk --min-score 7.5    # Filter by minimum score`,
	Run: func(cmd *cobra.Command, args []string) {
		runSeasonalBulk()
	},
}

var (
	bulkAddAll      bool
	bulkMinScore    float64
	bulkAutoConfirm bool
)

func runSeasonalBulk() {
	fmt.Println("🔍 Fetching current season anime...")
	animes, err := utils.FetchCurrentlyAiringSeason()
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}

	// Apply filters
	if seasonalFilter || bulkMinScore > 0 {
		var filtered []models.AnimeData
		for _, a := range animes {
			if a.Status == "Currently Airing" && a.Score >= bulkMinScore {
				filtered = append(filtered, a)
			}
		}
		animes = filtered
		fmt.Printf("🎬 Filtered to %d anime (score ≥ %.1f, currently airing)\n",
			len(animes), bulkMinScore)
	}

	if len(animes) == 0 {
		fmt.Println("⚠️  No anime match your filters.")
		return
	}

	var selected []models.AnimeData

	if bulkAddAll {
		selected = animes
	} else {
		selected = runMultiSelectMenu(animes)
		if len(selected) == 0 {
			fmt.Println("⚠️  No anime selected.")
			return
		}
	}

	// Confirmation
	if !bulkAutoConfirm {
		fmt.Printf("\n📋 Selected %d anime:\n", len(selected))
		for _, a := range selected {
			fmt.Printf("  • %s (%.1f)\n", a.Title, a.Score)
		}
		confirm := promptui.Prompt{
			Label:     "Add these to your watchlist?",
			IsConfirm: true,
		}
		if _, err := confirm.Run(); err != nil {
			fmt.Println("Cancelled.")
			return
		}
	}

	// Bulk add
	added, skipped, err := utils.BulkAddAnimeToList(selected)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	fmt.Printf("\n✅ Bulk add complete: %d added, %d skipped (duplicates)\n", added, skipped)
}

// runMultiSelectMenu provides a simplified multi-select interface
func runMultiSelectMenu(animes []models.AnimeData) []models.AnimeData {
	type selectable struct {
		models.AnimeData
		selected bool
		display  string
	}

	items := make([]selectable, len(animes))
	for i, a := range animes {
		items[i] = selectable{
			AnimeData: a,
			selected:  false,
			display:   fmt.Sprintf("%s (%.1f) [%s]", a.Title, a.Score, a.Status),
		}
	}
	items = append(items, selectable{
		AnimeData: models.AnimeData{Title: "✅ Confirm Selection"},
		display:   "✅ Confirm & Add Selected",
	})

	for {
		prompt := promptui.Select{
			Label:    "Select anime (navigate with arrows, Enter to toggle)",
			Items:    items,
			HideHelp: true,
			Templates: &promptui.SelectTemplates{
				Active:   `{{ if .selected }}[✓]{{ else }}[ ]{{ end }} {{ .display | cyan }}`,
				Inactive: `{{ if .selected }}[✓]{{ else }}[ ]{{ end }} {{ .display }}`,
				Selected: `→ {{ .display | green }}`,
			},
		}

		index, _, err := prompt.Run()
		if err != nil {
			return nil
		}
		// ✨ Clear the prompt label line
		fmt.Print("\033[1A\033[2K\r")

		// Confirm action
		if index == len(animes) {
			var result []models.AnimeData
			for _, item := range items[:len(animes)] {
				if item.selected {
					result = append(result, item.AnimeData)
				}
			}
			return result
		}

		// Toggle selection
		items[index].selected = !items[index].selected
		status := map[bool]string{true: "✓ Selected", false: "✗ Deselected"}[items[index].selected]
		fmt.Printf("  %s: %s\n", items[index].Title, status)
	}
}

func init() {
	seasonalCmd.AddCommand(seasonalBulkCmd)
	seasonalBulkCmd.Flags().BoolVarP(&bulkAddAll, "all", "a", false, "Add all currently airing seasonal anime")
	seasonalBulkCmd.Flags().Float64VarP(&bulkMinScore, "min-score", "m", 0, "Minimum score filter (0-10)")
	seasonalBulkCmd.Flags().BoolVarP(&bulkAutoConfirm, "yes", "y", false, "Skip confirmation prompt")
}
