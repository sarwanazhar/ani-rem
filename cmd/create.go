package cmd

import (
	"ani-rem/utils"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var name string

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new anime in the list",
	Run: func(cmd *cobra.Command, args []string) {
		if name == "" {
			prompt := promptui.Prompt{
				Label: "Enter Anime Name",
			}
			result, err := prompt.Run()
			if err != nil {
				if err == promptui.ErrInterrupt {
					os.Exit(0)
				}
				fmt.Printf("Prompt failed: %v\n", err)
				return
			}
			name = result
		}

		results, err := utils.SearchAnime(name)
		if err != nil {
			fmt.Printf("Search error: %v\n", err)
			return
		}
		if len(results) == 0 {
			fmt.Println("No results found for that name.")
			return
		}

		templates := &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "➤ {{ .Title | cyan }} ({{ .Score | yellow }})",
			Inactive: "  {{ .Title | white }}",
			Selected: "✔ {{ .Title | green }}",
		}

		prompt := promptui.Select{
			Label:     "Select Anime",
			Items:     results,
			Templates: templates,
		}

		i, _, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				os.Exit(0)
			}
			fmt.Println("Selection cancelled.")
			return
		}
		if i < 0 || i >= len(results) {
			fmt.Println("Invalid selection.")
			return
		}
		selectedAnime := results[i]

		actionPrompt := promptui.Select{
			Label: "What would you like to do with " + selectedAnime.Title + "?",
			Items: []string{"Confirm & Add to List", "Show Details", "Do Nothing"},
		}
		_, action, err := actionPrompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				os.Exit(0)
			}
			fmt.Println("Action cancelled.")
			return
		}

		switch action {
		case "Confirm & Add to List":
			err := utils.SaveAnimeToList(selectedAnime)
			if err != nil {
				fmt.Println("Error saving:", err)
			} else {
				fmt.Println("🚀 Added to your watch list!")
			}
		case "Show Details":
			fmt.Printf("\n--- %s ---\n", selectedAnime.Title)
			fmt.Printf("Status: %s | Score: %.2f\n", selectedAnime.Status, selectedAnime.Score)
			fmt.Println("\nSynopsis:", selectedAnime.Synopsis)

			confirmPrompt := promptui.Prompt{
				Label:     "Add to list now? (y/N)",
				IsConfirm: true,
			}
			if _, err := confirmPrompt.Run(); err == nil {
				utils.SaveAnimeToList(selectedAnime)
				fmt.Println("🚀 Added to your watch list!")
			}
		case "Do Nothing":
			fmt.Println("Action cancelled.")
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the resource")
}
