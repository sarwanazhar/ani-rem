package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func SendNotification(name string, Time string) {

	SendNotificationLogic(name, Time)
}

// ------------------------------------------
// ------------------------------------------
// ------------------------------------------
// ------------------------------------------

func ShouldSendNotification(name string) bool {
	// 1. Sanitize the name to use as a filename
	safeName := strings.ReplaceAll(name, " ", "_")
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("notify_%s.lock", safeName))

	fileInfo, err := os.Stat(tmpFile)

	// If file doesn't exist, we should definitely send it
	if os.IsNotExist(err) {
		return true
	}

	// 2. Check if the file is older than 2 hours
	// If the difference between NOW and the LAST MODIFIED time is > 2h
	if time.Since(fileInfo.ModTime()) > 1*time.Hour {
		return true
	}

	// Otherwise, it's too soon
	return false
}

func MarkAsSent(name string) {
	safeName := strings.ReplaceAll(name, " ", "_")
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("notify_%s.lock", safeName))

	// Create or update the timestamp of the file
	_ = os.WriteFile(tmpFile, []byte("sent"), 0644)
}

func SendNotificationLogic(name string, Time string) {
	if ShouldSendNotification(name) {
		fmt.Printf("Sending notification for: %s\n", name)

		// Code to sent notification

		// Create dynamic strings using fmt.Sprintf
		title := fmt.Sprintf("Alert for %s", name)
		message := fmt.Sprintf("Critical update: %s anime releasing soon in %s", name, Time)

		// -u critical: Sets urgency to high
		// -t 0: Sets timeout to 0 (won't expire)
		cmd := exec.Command("notify-send",
			"-u", "critical",
			"-t", "0",
			title,   // Using the dynamic title
			message, // Using the dynamic message
		)

		// Keep environment variables for DBUS/Display access
		cmd.Env = append(os.Environ(), "DISPLAY=:0", "DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus")

		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error sending notification for %s: %s\n", name, err)
		}

		MarkAsSent(name)
	} else {
		fmt.Printf("Skipping: Notification for %s was sent recently.\n", name)
	}
}
