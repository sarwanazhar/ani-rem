package cmd

import (
	"ani-rem/models"
	"ani-rem/utils"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the background reminder worker",
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Check if we are already a background process
		if os.Getenv("ANI_REM_CHILD") != "1" {
			// Start the same command again but as a detached child
			child := exec.Command(os.Args[0], "start")
			child.Env = append(os.Environ(), "ANI_REM_CHILD=1")

			err := child.Start()
			if err != nil {
				fmt.Printf("Failed to start background worker: %v\n", err)
				return
			}

			// Save the PID so we can stop it later
			_ = os.WriteFile(os.TempDir()+"/ani-rem.pid", []byte(fmt.Sprintf("%d", child.Process.Pid)), 0644)

			fmt.Println("🚀 Background worker started successfully!")
			fmt.Println("You can now close this terminal.")
			os.Exit(0)
		}

		// 2. The Worker Loop (This runs in the background)
		for {
			checkAiringAnime()
			time.Sleep(1 * time.Minute)
		}
	},
}

// stopCmd to kill the background process
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the background reminder worker",
	Run: func(cmd *cobra.Command, args []string) {
		pidData, err := os.ReadFile(os.TempDir() + "/ani-rem.pid")
		if err != nil {
			fmt.Println("No active worker found (or PID file missing).")
			return
		}

		fmt.Printf("Stopping worker (PID %s)...\n", string(pidData))

		// On Linux, we use the 'kill' command
		killCmd := exec.Command("kill", string(pidData))
		err = killCmd.Run()

		if err == nil {
			os.Remove(os.TempDir() + "/ani-rem.pid")
			fmt.Println("🛑 Worker stopped.")
		} else {
			fmt.Println("Failed to stop worker. It might have already exited.")
		}
	},
}

func checkAiringAnime() {
	filePath := utils.GetStoragePath()
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	var animes []models.AnimeData
	json.Unmarshal(fileData, &animes)

	for _, anime := range animes {
		if anime.Status == "Currently Airing" {
			remaining := utils.GetTimeUntilAiring(anime.Status, anime.Broadcast.Time, anime.Broadcast.Day)

			// Logic for different notification windows
			if strings.Contains(remaining, "0h 30m") {
				utils.SendNotification("30 Mins Left!", anime.Title+" airs in 30 minutes.")
			} else if strings.Contains(remaining, "1h 0m") {
				utils.SendNotification("1 Hour Left!", anime.Title+" airs in 1 hour.")
			} else if strings.Contains(remaining, "24h 0m") {
				utils.SendNotification("Airing Tomorrow", anime.Title+" airs in exactly 1 day.")
			}
		}
	}
}

var rootCmd = &cobra.Command{
	Use:   "ani-rem",
	Short: "ani-rem - Your CLI Anime Reminder & Watchlist",
	Run: func(cmd *cobra.Command, args []string) {
		for {
			prompt := promptui.Select{
				Label: "Main Menu",
				Items: []string{"Search & Add Anime", "View My Watchlist", "Start Background Worker", "Stop Background Worker", "Exit"},
			}

			_, result, err := prompt.Run()
			if err != nil {
				if err == promptui.ErrInterrupt {
					os.Exit(0)
				}
				return
			}

			switch result {
			case "Search & Add Anime":
				createCmd.Run(createCmd, args)
			case "View My Watchlist":
				listCmd.Run(listCmd, args)
			case "Start Background Worker":
				startCmd.Run(startCmd, args)
			case "Stop Background Worker":
				stopCmd.Run(stopCmd, args)
			case "Exit":
				os.Exit(0)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
