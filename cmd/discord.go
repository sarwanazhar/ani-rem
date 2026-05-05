package cmd

import (
	"ani-rem/utils"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var discordCmd = &cobra.Command{
	Use:   "discord",
	Short: "Manage Discord webhook notifications",
	Run: func(cmd *cobra.Command, args []string) {
		runDiscordMenu()
	},
}

func runDiscordMenu() {
	for {
		cfg, _ := utils.LoadConfig()
		status := "❌ Disabled"
		if cfg.DiscordEnabled {
			status = "✅ Enabled"
		}
		webhookDisplay := utils.MaskWebhookURL(cfg.DiscordWebhookURL)

		prompt := promptui.Select{
			Label: fmt.Sprintf("🎌 Discord Notifications [%s] | %s", status, webhookDisplay),
			Items: []string{
				"🔗 Setup / Update Webhook URL",
				"🔔 Toggle Notifications",
				"🧪 Test Connection",
				"↩️  Back to main menu",
			},
			Templates: &promptui.SelectTemplates{
				Active:   "➤ {{ . | cyan }}",
				Inactive: "  {{ . }}",
				Selected: "✔ {{ . | green }}",
			},
		}

		idx, _, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				os.Exit(0)
			}
			return
		}

		switch idx {
		case 0:
			setupWebhook()
		case 1:
			toggleDiscord()
		case 2:
			testConnection()
		case 3:
			return
		}
	}
}

func setupWebhook() {
	fmt.Println("\n🔗 Discord Webhook Setup")
	fmt.Println(strings.Repeat("=", 40))
	fmt.Println("\n1. Create Your Private Server")
	fmt.Println("   • Open Discord and click the '+' (Add a Server) icon")
	fmt.Println("   • Choose 'Create My Own' → 'For me and my friends'")
	fmt.Println("   • Name it something like 'Ani-Rem Alerts' and click Create")
	fmt.Println("\n2. Create a Private Channel")
	fmt.Println("   • In your new server, click '+' next to 'Text Channels'")
	fmt.Println("   • Name the channel 'updates'")
	fmt.Println("   • (Optional) Make it Private if you want only you to see it")
	fmt.Println("\n3. Generate the Webhook URL")
	fmt.Println("   • Hover over #updates → click Settings (Gear Icon)")
	fmt.Println("   • Click 'Integrations' → 'Webhooks' → 'New Webhook'")
	fmt.Println("   • Click the new webhook → rename to 'Ani-Rem Bot' (optional)")
	fmt.Println("   • Click 'Copy Webhook URL' and paste it below")
	fmt.Println()

	webhookPrompt := promptui.Prompt{
		Label: "Paste your Discord Webhook URL",
		Validate: func(input string) error {
			input = strings.TrimSpace(input)
			if input == "" {
				return fmt.Errorf("webhook URL cannot be empty")
			}
			if !isValidWebhookURL(input) {
				return fmt.Errorf("invalid Discord webhook URL format")
			}
			return nil
		},
	}

	webhookURL, err := webhookPrompt.Run()
	if err != nil {
		fmt.Println("❌ Setup cancelled.")
		return
	}

	cfg, _ := utils.LoadConfig()
	cfg.DiscordWebhookURL = strings.TrimSpace(webhookURL)
	if err := utils.SaveConfig(cfg); err != nil {
		fmt.Printf("❌ Failed to save webhook URL: %v\n", err)
		return
	}

	fmt.Println("\n✅ Webhook URL saved successfully!")
	fmt.Printf("   Stored as: %s\n", utils.MaskWebhookURL(cfg.DiscordWebhookURL))
	fmt.Println("\n💡 Tip: Run 'Test Connection' to verify it works.")
}

func isValidWebhookURL(u string) bool {
	u = strings.TrimSpace(u)
	if !strings.HasPrefix(u, "https://discord.com/api/webhooks/") &&
		!strings.HasPrefix(u, "https://discordapp.com/api/webhooks/") {
		return false
	}
	// Basic pattern: /webhooks/{id}/{token}
	parts := strings.Split(strings.TrimSuffix(u, "/"), "/")
	if len(parts) < 2 {
		return false
	}
	id := parts[len(parts)-2]
	token := parts[len(parts)-1]
	// IDs and tokens are alphanumeric + underscore/hyphen, 60-80 chars typical
	idMatch, _ := regexp.MatchString(`^[a-zA-Z0-9_-]{10,}$`, id)
	tokenMatch, _ := regexp.MatchString(`^[a-zA-Z0-9_-]{40,}$`, token)
	return idMatch && tokenMatch
}

func toggleDiscord() {
	cfg, err := utils.LoadConfig()
	if err != nil {
		fmt.Printf("⚠️  Could not load config: %v\n", err)
		return
	}

	if cfg.DiscordWebhookURL == "" {
		fmt.Println("\n⚠️  No webhook URL configured.")
		fmt.Println("   Run 'Setup / Update Webhook URL' first.")
		return
	}

	newState := !cfg.DiscordEnabled
	cfg.DiscordEnabled = newState

	if err := utils.SaveConfig(cfg); err != nil {
		fmt.Printf("❌ Failed to save: %v\n", err)
		return
	}

	status := map[bool]string{true: "✅ Enabled", false: "❌ Disabled"}[newState]
	fmt.Printf("\n🔔 Discord notifications %s.\n", status)
	if newState {
		fmt.Println("   You will now receive alerts in your Discord channel.")
	} else {
		fmt.Println("   Desktop notifications will continue working as before.")
	}
}

func testConnection() {
	cfg, err := utils.LoadConfig()
	if err != nil {
		fmt.Printf("⚠️  Could not load config: %v\n", err)
		return
	}

	webhookURL := strings.TrimSpace(cfg.DiscordWebhookURL)
	if webhookURL == "" {
		fmt.Println("\n❌ No webhook URL configured.")
		fmt.Println("   Run 'Setup / Update Webhook URL' first.")
		return
	}

	// Step 1: Validate format locally
	if !isValidWebhookURL(webhookURL) {
		fmt.Println("\n❌ Invalid webhook URL format.")
		fmt.Println("   Please re-run 'Setup / Update Webhook URL' with a valid URL.")
		return
	}
	fmt.Println("\n✅ Webhook URL format is valid.")

	// Step 2: Attempt HEAD request to verify reachability
	fmt.Print("🔍 Verifying endpoint reachability... ")
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("HEAD", webhookURL, nil)
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		return
	}
	req.Header.Set("User-Agent", "ani-rem/1.0")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		fmt.Println("\n💡 Possible causes:")
		fmt.Println("   • Invalid or expired webhook URL")
		fmt.Println("   • Network/firewall blocking Discord API")
		fmt.Println("   • Discord API temporarily unavailable")
		return
	}
	defer resp.Body.Close()

	// Discord webhooks return 200 OK or 204 No Content for valid endpoints
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		fmt.Println("✅ Endpoint is reachable!")
		fmt.Println("\n🎉 Your Discord integration is ready.")
		fmt.Println("   Run 'Toggle Notifications' to enable alerts.")
	} else {
		fmt.Printf("⚠️  Endpoint returned status: %d %s\n", resp.StatusCode, resp.Status)
		fmt.Println("\n💡 The URL format is valid, but Discord rejected the request.")
		fmt.Println("   This may mean the webhook was deleted or revoked.")
		fmt.Println("   Try re-generating the webhook URL in Discord.")
	}
}

func init() {
	rootCmd.AddCommand(discordCmd)
}
