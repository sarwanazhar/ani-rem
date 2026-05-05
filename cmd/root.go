package cmd

import (
	"ani-rem/models"
	"ani-rem/utils"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var autoSync bool

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the background reminder worker",
	Run: func(cmd *cobra.Command, args []string) {
		if os.Getenv("ANI_REM_CHILD") != "1" {
			// Parent: launch child and wait for PID file
			childArgs := []string{"start"}
			if autoSync {
				childArgs = append(childArgs, "--auto-sync")
			}
			child := exec.Command(os.Args[0], childArgs...)
			child.Env = append(os.Environ(), "ANI_REM_CHILD=1")

			if err := child.Start(); err != nil {
				fmt.Printf("Failed to start background worker: %v\n", err)
				return
			}

			// Wait up to 2 seconds for child to write its PID file
			pidFile := os.TempDir() + "/ani-rem.pid"
			var pidData []byte
			for i := 0; i < 20; i++ {
				time.Sleep(100 * time.Millisecond)
				pidData, _ = os.ReadFile(pidFile)
				if len(pidData) > 0 {
					break
				}
			}
			if len(pidData) == 0 {
				fmt.Println("⚠️ Background worker started but PID file not detected.")
				fmt.Println("Worker may still be running. Check with 'ani-rem stop' later.")
				return
			}

			// Verify the process exists
			pidStr := string(pidData)
			checkCmd := exec.Command("kill", "-0", pidStr)
			if err := checkCmd.Run(); err != nil {
				fmt.Printf("⚠️ Worker started but process %s is not responding.\n", pidStr)
				return
			}

			fmt.Println("🚀 Background worker started successfully!")
			fmt.Println("You can now close this terminal.")
			os.Exit(0)
		}

		// Child: set up signal handling for graceful shutdown
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		go func() {
			<-sigCh
			os.Remove(os.TempDir() + "/ani-rem.pid")
			os.Exit(0)
		}()

		// Write PID file with restrictive permissions (only owner readable)
		pid := os.Getpid()
		if err := os.WriteFile(os.TempDir()+"/ani-rem.pid", []byte(fmt.Sprintf("%d", pid)), 0600); err != nil {
			fmt.Printf("Warning: could not write PID file: %v\n", err)
		}

		fmt.Println("Background worker running...")
		for {
			utils.CheckAiringAnime()
			// Use config flag if auto‑sync is enabled at start or via config
			if autoSync || isAutoSyncEnabled() {
				syncOnceADay()
			}
			time.Sleep(5 * time.Minute)
		}
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the background reminder worker",
	Run: func(cmd *cobra.Command, args []string) {
		pidData, err := os.ReadFile(os.TempDir() + "/ani-rem.pid")
		if err != nil {
			fmt.Println("No active worker found (or PID file missing).")
			return
		}
		pid := string(pidData)
		fmt.Printf("Stopping worker (PID %s)...\n", pid)
		killCmd := exec.Command("kill", pid)
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
				Items: []string{
					"Search & Add Anime",
					"📺 Browse Seasonal Anime",
					"View My Watchlist",
					"Start Background Worker",
					"Stop Background Worker",
					"📅 Google Calendar",
					"⚙️  Settings",
					"Exit",
				},
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
			case "📺 Browse Seasonal Anime": // ← NEW
				seasonalCmd.Run(seasonalCmd, args)
			case "View My Watchlist":
				listCmd.Run(listCmd, args)
			case "Start Background Worker":
				startCmd.Run(startCmd, args)
			case "Stop Background Worker":
				stopCmd.Run(stopCmd, args)
			case "📅 Google Calendar":
				calendarCmd.Run(calendarCmd, args)
			case "⚙️  Settings":
				configCmd.Run(configCmd, args)
			case "Exit":
				os.Exit(0)
			}
		}
	},
}

var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ani-rem version %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	startCmd.Flags().BoolVar(&autoSync, "auto-sync", false, "Automatically sync anime schedule to Google Calendar once per day")
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// isAutoSyncEnabled checks config, returns true if auto‑sync is on.
func isAutoSyncEnabled() bool {
	cfg, err := utils.LoadConfig()
	if err != nil {
		return false
	}
	return cfg.AutoSync
}

func syncOnceADay() {
	lastSyncFile := filepath.Join(os.TempDir(), "ani-rem-last-sync")

	info, err := os.Stat(lastSyncFile)
	if err == nil {
		lastSync := info.ModTime()
		now := time.Now()
		if now.Year() == lastSync.Year() && now.YearDay() == lastSync.YearDay() {
			return
		}
	}

	fmt.Println("🔄 Auto-syncing anime schedule to Google Calendar...")

	filePath := utils.GetStoragePath()
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Auto-sync: could not read anime list")
		return
	}

	var animes []models.AnimeData
	if err := json.Unmarshal(fileData, &animes); err != nil {
		fmt.Println("Auto-sync: error parsing anime list:", err)
		return
	}

	var airing []models.AnimeData
	for _, a := range animes {
		if a.Status == "Currently Airing" {
			airing = append(airing, a)
		}
	}
	if len(airing) == 0 {
		fmt.Println("Auto-sync: no currently airing anime found")
		return
	}

	client, err := utils.NewGoogleCalendarClient()
	if err != nil || !client.IsAuthenticated() {
		fmt.Println("Auto-sync: not authenticated, skipping. Run 'ani-rem calendar connect' first.")
		return
	}

	cfg, _ := utils.LoadConfig()
	calendarID := cfg.CalendarID
	if calendarID == "" {
		id, err := client.GetPrimaryCalendarID()
		if err != nil {
			fmt.Printf("Auto-sync: cannot get primary calendar: %v\n", err)
			return
		}
		calendarID = id
	}

	err = client.SyncMultipleAnime(airing, 12, calendarID)
	if err != nil {
		fmt.Printf("Auto-sync: sync failed: %v\n", err)
		return
	}

	_ = os.WriteFile(lastSyncFile, []byte(time.Now().String()), 0644)
	fmt.Println("Auto-sync completed.")
}
