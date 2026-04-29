package utils

import (
	"fmt"
	"os"
	"os/exec"
)

func SendNotification(title, message string) {
	// 1. Prepare the command with your custom flags
	// -u critical: Makes the notification stay on screen until clicked
	// -i input-tablet: Uses the tablet icon you requested
	cmd := exec.Command("notify-send", "-u", "critical", "-i", "input-tablet", title, message)

	// 2. Crucial for background processes: Tell it which display to use
	// On most Linux Mint setups, this is :0
	cmd.Env = append(os.Environ(), "DISPLAY=:0", "DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus")

	// 3. Run it
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error sending notification:", err)
	}
}
