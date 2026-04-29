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
		// If flag is empty, prompt the user
		if name == "" {
			prompt := promptui.Prompt{
				Label: "Enter Anime Name",
			}

			result, err := prompt.Run()
			if err != nil {
				if err == promptui.ErrInterrupt {
					os.Exit(0)
				}
				fmt.Printf("Prompt failed %v\n", err)
				return
			}
			name = result
		}

		// Inside your Command Run function
		results, _ := utils.SearchAnime(name)

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

		i, _, _ := prompt.Run()
		selectedAnime := results[i]

		// 1. Create the action menu
		actionPrompt := promptui.Select{
			Label: "What would you like to do with " + selectedAnime.Title + "?",
			Items: []string{"Confirm & Add to List", "Show Details", "Do Nothing"},
		}

		_, action, err := actionPrompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				os.Exit(0)
			}
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

			// Recursive call or simple prompt to add after reading
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

		// YOUR CODE GOES HERE:
		// This is where you start writing your custom logic!
	},
}

func init() {
	// Attach the command to the root
	rootCmd.AddCommand(createCmd)

	// Define the flag
	createCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the resource")
}
