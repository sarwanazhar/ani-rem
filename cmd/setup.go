package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var setupCalendarCmd = &cobra.Command{
	Use:   "setup-calendar",
	Short: "Guided setup for Google Calendar integration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📅 Google Calendar Integration Setup")
		fmt.Println("===================================")
		fmt.Println()
		fmt.Println("Follow these steps to connect ani-rem to your Google Calendar:")
		fmt.Println()
		fmt.Println("Step 1: Create a Google Cloud Project")
		fmt.Println("   → Visit: https://console.cloud.google.com/")
		fmt.Println("   → Click 'Create Project'")
		fmt.Println("   → Name: ani-rem")
		fmt.Println()
		fmt.Println("Step 2: Enable Google Calendar API")
		fmt.Println("   → In your project, go to 'APIs & Services' → 'Library'")
		fmt.Println("   → Search for 'Google Calendar API'")
		fmt.Println("   → Click 'Enable'")
		fmt.Println()
		fmt.Println("Step 3: Configure OAuth Consent Screen")
		fmt.Println("   → Go to 'APIs & Services' → 'OAuth consent screen'")
		fmt.Println("   → User Type: 'External'")
		fmt.Println("   → Fill in:")
		fmt.Println("      • App name: ani-rem")
		fmt.Println("      • User support email: your email")
		fmt.Println("      • Developer contact: your email")
		fmt.Println("   → Click 'Save and Continue'")
		fmt.Println("   → Add scope: .../auth/calendar (OR .../auth/calendar.events)")
		fmt.Println("   → Add test users: your email address")
		fmt.Println("   → Save")
		fmt.Println()
		fmt.Println("Step 4: Create Credentials")
		fmt.Println("   → Go to 'APIs & Services' → 'Credentials'")
		fmt.Println("   → Click '+ Create Credentials' → 'OAuth client ID'")
		fmt.Println("   → Application type: 'Desktop app'")
		fmt.Println("   → Name: ani-rem desktop")
		fmt.Println("   → Click 'Create'")
		fmt.Println()
		fmt.Println("Step 5: Copy Client ID and Secret")
		fmt.Println("   → You will see a dialog with Client ID and Client Secret")
		fmt.Println("   → Copy both (they will be used in the next step)")
		fmt.Println()
		fmt.Println("Step 6: Run 'ani-rem calendar connect' and paste them when prompted")
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(setupCalendarCmd)
}
