package cmd

import (
	"ani-rem/utils"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage ani-rem settings",
	Run: func(cmd *cobra.Command, args []string) {
		for {
			prompt := promptui.Select{
				Label: "⚙️  Settings",
				Items: []string{
					"📄 View current settings",
					"✏️  Edit settings in $EDITOR",
					"🔄 Auto-sync toggle",
					"⏰ Change notification threshold",
					"↩️  Back to main menu",
				},
			}
			idx, _, err := prompt.Run()
			if err != nil {
				return
			}
			switch idx {
			case 0:
				viewConfig()
			case 1:
				editConfig()
			case 2:
				toggleAutoSync()
			case 3:
				changeThreshold()
			case 4:
				return
			}
		}
	},
}

func viewConfig() {
	cfg, err := utils.LoadConfig()
	if err != nil {
		fmt.Printf("⚠️  %v\n", err)
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Println("\n" + string(data))
	fmt.Print("\nPress Enter to continue...")
	fmt.Scanln()
}

func editConfig() {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano" // fallback
	}

	path := utils.GetConfigPath()
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Editor failed: %v\n", err)
	} else {
		fmt.Println("✅ Settings saved.")
	}
}

func toggleAutoSync() {
	cfg, _ := utils.LoadConfig()
	cfg.AutoSync = !cfg.AutoSync
	if err := utils.SaveConfig(cfg); err != nil {
		fmt.Printf("❌ Failed to save: %v\n", err)
		return
	}
	status := "off"
	if cfg.AutoSync {
		status = "on"
	}
	fmt.Printf("✅ Auto‑sync turned %s.\n", status)
}

func changeThreshold() {
	cfg, _ := utils.LoadConfig()
	prompt := promptui.Prompt{
		Label:   "Notification threshold in hours (current: " + fmt.Sprint(cfg.NotificationThresholdHours) + ")",
		Default: fmt.Sprint(cfg.NotificationThresholdHours),
		Validate: func(input string) error {
			value, err := strconv.Atoi(input)
			if err != nil || value <= 0 {
				return fmt.Errorf("enter a positive integer")
			}
			return nil
		},
	}
	result, err := prompt.Run()
	if err != nil {
		return
	}
	hours, _ := strconv.Atoi(result)
	cfg.NotificationThresholdHours = hours
	if err := utils.SaveConfig(cfg); err != nil {
		fmt.Printf("❌ Failed to save: %v\n", err)
		return
	}
	fmt.Printf("✅ Notification threshold set to %d hours.\n", hours)
}

func init() {
	rootCmd.AddCommand(configCmd)
}
