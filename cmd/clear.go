package cmd

import (
	"ani-rem/utils"
	"fmt"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"google.golang.org/api/calendar/v3"
)

var forceClear bool

var clearCalendarCmd = &cobra.Command{
	Use:   "clear",
	Short: "Delete all anime events from Google Calendar (created by ani-rem)",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := utils.NewGoogleCalendarClient()
		if err != nil {
			fmt.Printf("❌ Not connected: %v\n", err)
			fmt.Println("Run 'ani-rem calendar connect' first.")
			return
		}
		if !client.IsAuthenticated() {
			fmt.Println("❌ Not authenticated. Run 'ani-rem calendar connect'.")
			return
		}

		calendarID, err := client.GetPrimaryCalendarID()
		if err != nil {
			fmt.Printf("❌ Failed to get primary calendar: %v\n", err)
			return
		}

		// List events created by ani-rem
		fmt.Println("🔍 Searching for anime events created by ani-rem...")
		events, err := findAnimeEvents(client, calendarID)
		if err != nil {
			fmt.Printf("❌ Failed to search events: %v\n", err)
			return
		}

		if len(events) == 0 {
			fmt.Println("✅ No anime events found in your calendar.")
			return
		}

		fmt.Printf("\n📋 Found %d anime event(s):\n", len(events))
		// Show first 10 as preview to avoid clutter
		showCount := 10
		if len(events) < showCount {
			showCount = len(events)
		}
		for i := 0; i < showCount; i++ {
			ev := events[i]
			start := ev.Start.DateTime
			if start == "" {
				start = ev.Start.Date
			}
			fmt.Printf("  %d. %s (%s)\n", i+1, ev.Summary, start)
		}
		if len(events) > showCount {
			fmt.Printf("  ... and %d more\n", len(events)-showCount)
		}
		fmt.Println()

		if !forceClear {
			confirm := promptui.Prompt{
				Label:     fmt.Sprintf("Delete all %d events permanently", len(events)),
				IsConfirm: true,
			}
			if _, err := confirm.Run(); err != nil {
				fmt.Println("Clear cancelled.")
				return
			}
		}

		// Delete events sequentially (respects Google API rate limits)
		success := 0
		fail := 0
		for _, ev := range events {
			err := client.DeleteEvent(calendarID, ev.Id)
			if err != nil {
				fmt.Printf("  ❌ Failed to delete %s: %v\n", ev.Summary, err)
				fail++
			} else {
				fmt.Printf("  ✓ Deleted: %s\n", ev.Summary)
				success++
			}
		}
		fmt.Printf("\n📊 Deleted %d events, %d failed.\n", success, fail)
	},
}

// findAnimeEvents returns all events whose summary contains "📺" and "- New Episode"
// or description contains "Powered by ani-rem"
func findAnimeEvents(client *utils.GoogleCalendarClient, calendarID string) ([]*calendar.Event, error) {
	now := time.Now()
	timeMin := now.AddDate(-1, 0, 0).Format(time.RFC3339)
	timeMax := now.AddDate(2, 0, 0).Format(time.RFC3339)

	events, err := client.ListEvents(calendarID, timeMin, timeMax)
	if err != nil {
		return nil, err
	}

	var animeEvents []*calendar.Event
	for _, ev := range events {
		if strings.Contains(ev.Summary, "📺") && strings.Contains(ev.Summary, "- New Episode") {
			animeEvents = append(animeEvents, ev)
			continue
		}
		if strings.Contains(ev.Description, "Powered by ani-rem") {
			animeEvents = append(animeEvents, ev)
		}
	}
	return animeEvents, nil
}

func init() {
	calendarCmd.AddCommand(clearCalendarCmd)
	clearCalendarCmd.Flags().BoolVarP(&forceClear, "force", "f", false, "Delete without confirmation")
}
