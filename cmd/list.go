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

			displayItems := append(animes, models.AnimeData{Title: "🗑️  Delete Entire List"})
			displayItems = append(displayItems, models.AnimeData{Title: "➜ Exit to Menu"})

			prompt := promptui.Select{
				Label: "Your Watchlist",
				Items: displayItems,
				Templates: &promptui.SelectTemplates{
					Active:   "➤ {{ .Title | cyan }}",
					Inactive: "  {{ .Title }}",
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

			// Exit to Menu
			if index == len(displayItems)-1 {
				return
			}

			// Delete Entire List
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
			selected := animes[index]

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
				fmt.Println("\n(Press Enter to go back to list)")
				fmt.Scanln()
				continue
			}

			if action == "Delete from List" {
				animes = append(animes[:index], animes[index+1:]...)
				if err := utils.UpdateFullList(animes); err != nil {
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
