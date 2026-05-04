package cmd

import (
	"ani-rem/utils"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var calendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "Connect and manage Google Calendar integration",
	Run: func(cmd *cobra.Command, args []string) {
		for {
			prompt := promptui.Select{
				Label: "Google Calendar Integration",
				Items: []string{
					"🔐 Connect / Sign in to Google Calendar",
					"📅 List my calendars",
					"🔄 Sync anime to calendar",
					"🗑️  Clear all anime events",
					"❌ Remove specific anime events",
					"🚫 Disconnect (remove access)",
					"↩️  Back to main menu",
				},
			}
			index, _, err := prompt.Run()
			if err != nil {
				if err == promptui.ErrInterrupt {
					os.Exit(0)
				}
				return
			}
			switch index {
			case 0:
				connectToGoogleCalendar()
			case 1:
				listCalendars()
			case 2:
				syncCmd.Run(syncCmd, []string{})
			case 3:
				clearCalendarCmd.Run(clearCalendarCmd, []string{})
			case 4:
				removeSpecificCalendarCmd.Run(removeSpecificCalendarCmd, []string{})
			case 5:
				disconnectCalendar()
			case 6:
				return
			}
		}
	},
}

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to Google Calendar (prompts for credentials, then opens browser)",
	Run: func(cmd *cobra.Command, args []string) {
		connectToGoogleCalendar()
	},
}

var disconnectCalendarCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Remove Google Calendar authentication",
	Run: func(cmd *cobra.Command, args []string) {
		disconnectCalendar()
	},
}

func connectToGoogleCalendar() {
	fmt.Println("🔐 Connecting to Google Calendar...")

	client, err := utils.NewGoogleCalendarClient()
	if err == nil && client != nil && client.IsAuthenticated() {
		fmt.Println("✓ Already connected! You can start syncing with 'ani-rem sync'")
		return
	}

	fmt.Println("\nTo connect to Google Calendar, you need to provide OAuth 2.0 credentials.")
	fmt.Println("If you don't have them yet:")
	fmt.Println("1. Go to https://console.cloud.google.com/")
	fmt.Println("2. Create a new project or select existing")
	fmt.Println("3. Enable Google Calendar API")
	fmt.Println("4. Create OAuth 2.0 credentials (Desktop app type)")
	fmt.Println("5. Copy the Client ID and Client Secret")
	fmt.Println()

	clientIDPrompt := promptui.Prompt{
		Label: "Enter your Google OAuth Client ID",
		Validate: func(input string) error {
			if len(input) == 0 {
				return fmt.Errorf("Client ID cannot be empty")
			}
			return nil
		},
	}
	clientID, err := clientIDPrompt.Run()
	if err != nil {
		fmt.Println("❌ Cancelled")
		return
	}

	clientSecretPrompt := promptui.Prompt{
		Label: "Enter your Google OAuth Client Secret",
		Mask:  '*',
		Validate: func(input string) error {
			if len(input) == 0 {
				return fmt.Errorf("Client Secret cannot be empty")
			}
			return nil
		},
	}
	clientSecret, err := clientSecretPrompt.Run()
	if err != nil {
		fmt.Println("❌ Cancelled")
		return
	}

	client = utils.NewEmptyClient()
	if err := client.SetCredentials(clientID, clientSecret); err != nil {
		fmt.Printf("❌ Failed to save credentials: %v\n", err)
		return
	}

	err = client.Authenticate()
	if err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		return
	}

	fmt.Println("\n✓ Google Calendar is now connected!")
	fmt.Println("You can now run 'ani-rem sync' to add your anime schedule.")
}

func listCalendars() {
	client, err := utils.NewGoogleCalendarClient()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	if !client.IsAuthenticated() {
		fmt.Println("❌ Not connected to Google Calendar. Run 'ani-rem calendar connect' first.")
		return
	}
	err = client.ListCalendarLists()
	if err != nil {
		fmt.Printf("❌ Failed to list calendars: %v\n", err)
	}
}

func disconnectCalendar() {
	confirm := promptui.Prompt{
		Label:     "Are you sure you want to disconnect Google Calendar",
		IsConfirm: true,
	}
	if _, err := confirm.Run(); err != nil {
		fmt.Println("Disconnect cancelled.")
		return
	}
	tokenStore := utils.NewTokenStore()
	err := tokenStore.Delete()
	if err != nil {
		fmt.Printf("❌ Failed to disconnect: %v\n", err)
	} else {
		fmt.Println("✅ Disconnected from Google Calendar.")
		fmt.Println("Your anime events will remain in Google Calendar but won't be updated.")
	}
}

func init() {
	rootCmd.AddCommand(calendarCmd)
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(disconnectCalendarCmd)
}
