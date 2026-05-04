package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func SendNotification(name string, Time string) {
	SendNotificationLogic(name, Time)
}

func ShouldSendNotification(name string) bool {
	safeName := strings.ReplaceAll(name, " ", "_")
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("notify_%s.lock", safeName))

	fileInfo, err := os.Stat(tmpFile)
	if os.IsNotExist(err) {
		return true
	}
	if time.Since(fileInfo.ModTime()) > 1*time.Hour {
		return true
	}
	return false
}

func MarkAsSent(name string) {
	safeName := strings.ReplaceAll(name, " ", "_")
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("notify_%s.lock", safeName))
	_ = os.WriteFile(tmpFile, []byte("sent"), 0644)
}

func SendNotificationLogic(name string, Time string) {
	if ShouldSendNotification(name) {
		fmt.Printf("Sending notification for: %s\n", name)

		title := fmt.Sprintf("Alert for %s", name)
		message := fmt.Sprintf("Critical update: %s anime releasing soon in %s", name, Time)

		cmd := exec.Command("notify-send",
			"-u", "critical",
			"-t", "0",
			title,
			message,
		)

		display := os.Getenv("DISPLAY")
		if display == "" {
			display = ":0"
		}

		// Build D-Bus path dynamically using the current user's UID
		uid := os.Getuid()
		dbus := os.Getenv("DBUS_SESSION_BUS_ADDRESS")
		if dbus == "" {
			dbus = "unix:path=/run/user/" + strconv.Itoa(uid) + "/bus"
		}

		cmd.Env = append(os.Environ(), "DISPLAY="+display, "DBUS_SESSION_BUS_ADDRESS="+dbus)

		if err := cmd.Run(); err != nil {
			fmt.Printf("Error sending notification for %s: %v\n", name, err)
		}

		MarkAsSent(name)
	} else {
		fmt.Printf("Skipping: Notification for %s was sent recently.\n", name)
	}
}
