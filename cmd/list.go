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

// displayItem extends AnimeData with a calculated "Remaining" countdown.
type displayItem struct {
	models.AnimeData
	Remaining string
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "View and manage your saved anime",
	Run: func(cmd *cobra.Command, args []string) {
		for {
			filePath := utils.GetStoragePath()
			fileData, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Println("No saved anime found.")
				return
			}

			var animes []models.AnimeData
			if err := json.Unmarshal(fileData, &animes); err != nil {
				fmt.Printf("Error parsing anime list: %v\n", err)
				return
			}

			if len(animes) == 0 {
				fmt.Println("Your list is empty.")
				return
			}

			// Build a uniform slice of displayItem (no more mixed types)
			var items []displayItem
			for _, a := range animes {
				item := displayItem{AnimeData: a}
				if a.Status == "Currently Airing" {
					item.Remaining = utils.GetTimeUntilAiring(a.Status, a.Broadcast.Time, a.Broadcast.Day)
				}
				items = append(items, item)
			}
			// Append special entries as displayItem with only Title set
			items = append(items, displayItem{AnimeData: models.AnimeData{Title: "🗑️  Delete Entire List"}})
			items = append(items, displayItem{AnimeData: models.AnimeData{Title: "➜ Exit to Menu"}})

			prompt := promptui.Select{
				Label: "Your Watchlist",
				Items: items,
				Templates: &promptui.SelectTemplates{
					Label:    "{{ . }}",
					Active:   "➤ {{ .Title | cyan }}{{ if .Remaining }} {{ .Remaining | yellow }}{{ end }}",
					Inactive: "  {{ .Title }}{{ if .Remaining }} - {{ .Remaining }}{{ end }}",
					Selected: "✔ {{ .Title | green }}",
				},
			}

			index, _, err := prompt.Run()
			if err != nil {
				if err == promptui.ErrInterrupt {
					os.Exit(0)
				}
				return
			}

			// Handle special entries (they are appended after the anime)
			if index == len(animes)+1 {
				return // Exit to Menu
			}
			if index == len(animes) {
				confirmPrompt := promptui.Prompt{
					Label:     "Are you sure you want to delete the entire list file? (y/N)",
					IsConfirm: true,
				}
				if _, err := confirmPrompt.Run(); err == nil {
					err := os.Remove(filePath)
					if err != nil {
						fmt.Printf("Error deleting file: %v\n", err)
					} else {
						fmt.Println("🔥 List file deleted successfully.")
						return
					}
				} else {
					fmt.Println("Deletion cancelled.")
					continue
				}
			}

			// Selected a specific anime
			selected := items[index]

			actionPrompt := promptui.Select{
				Label: "Actions for " + selected.Title,
				Items: []string{"Show Details", "Delete from List", "Back"},
			}
			_, action, _ := actionPrompt.Run()

			if action == "Back" {
				continue
			}

			if action == "Show Details" {
				remaining := utils.GetTimeUntilAiring(selected.Status, selected.Broadcast.Time, selected.Broadcast.Day)

				fmt.Printf("\n--- %s ---\n", selected.Title)
				fmt.Printf("Status: %s\n", selected.Status)
				if selected.Status == "Currently Airing" {
					fmt.Printf("Next Airing: %s\n", remaining)
				}
				fmt.Println("\nSynopsis:", selected.Synopsis)

				// Use promptui's Prompt to pause without breaking terminal state
				pause := promptui.Prompt{
					Label:       "Press Enter to go back to list",
					AllowEdit:   false,
					HideEntered: true,
				}
				pause.Run()
				continue
			}

			if action == "Delete from List" {
				// Re‑read the raw list to delete precisely
				fileData, _ := os.ReadFile(filePath)
				var raw []models.AnimeData
				json.Unmarshal(fileData, &raw)
				raw = append(raw[:index], raw[index+1:]...)
				if err := utils.UpdateFullList(raw); err != nil {
					fmt.Printf("Error updating list: %v\n", err)
				} else {
					fmt.Printf("🗑️  Deleted %s.\n", selected.Title)
				}
				continue
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
