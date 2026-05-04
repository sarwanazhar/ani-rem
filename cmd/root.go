package cmd

import (
	"ani-rem/utils"
	"fmt"
	"os"
	"os/exec"
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

		for {
			utils.CheckAiringAnime()
			time.Sleep(5 * time.Minute)
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
