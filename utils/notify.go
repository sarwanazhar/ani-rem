package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
)

// sanitizeForFilename replaces any character that is unsafe for filenames.
func sanitizeForFilename(name string) string {
	// Keep only alphanumeric, underscores, hyphens, and dots.
	reg := regexp.MustCompile(`[^\w\.-]`)
	safe := reg.ReplaceAllString(name, "_")
	// Collapse multiple underscores and trim
	reg2 := regexp.MustCompile(`_+`)
	safe = reg2.ReplaceAllString(safe, "_")
	safe = strings.Trim(safe, "_")
	if safe == "" {
		safe = "unnamed"
	}
	return safe
}

func SendNotification(name string, Time string) {
	SendNotificationLogic(name, Time)
}

func ShouldSendNotification(name string) bool {
	safeName := sanitizeForFilename(name)
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
	safeName := sanitizeForFilename(name)
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("notify_%s.lock", safeName))
	_ = os.WriteFile(tmpFile, []byte("sent"), 0644)
}

func SendNotificationLogic(name string, Time string) {
	if ShouldSendNotification(name) {
		fmt.Printf("Sending notification for: %s\n", name)

		title := fmt.Sprintf("🎌 ani-rem: %s", name)
		message := fmt.Sprintf("Episode releasing soon in %s", Time)

		// Use beeep for cross-platform notifications
		err := beeep.Notify(title, message, "")
		if err != nil {
			fmt.Printf("Error sending notification for %s: %v\n", name, err)

			// Fallback to platform-specific commands if beeep fails
			fallbackNotification(title, message)
		}

		MarkAsSent(name)
	} else {
		fmt.Printf("Skipping: Notification for %s was sent recently.\n", name)
	}
}

// fallbackNotification provides platform-specific notification fallback
func fallbackNotification(title, message string) {
	switch runtime.GOOS {
	case "darwin":
		// macOS: use osascript
		cmd := exec.Command("osascript", "-e",
			fmt.Sprintf(`display notification "%s" with title "%s"`, message, title))
		cmd.Run()
	case "windows":
		// Windows: use PowerShell toast notification (Windows 10+)
		psScript := fmt.Sprintf(`
			[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
			[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null
			$template = @"
			<toast>
				<visual>
					<binding template="ToastText02">
						<text id="1">%s</text>
						<text id="2">%s</text>
					</binding>
				</visual>
			</toast>
"@
			$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
			$xml.LoadXml($template)
			$toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
			[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("ani-rem").Show($toast)
		`, title, message)
		cmd := exec.Command("powershell", "-Command", psScript)
		cmd.Run()
	case "linux":
		// Linux fallback: try notify-send with common display settings
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
		uid := os.Getuid()
		dbus := os.Getenv("DBUS_SESSION_BUS_ADDRESS")
		if dbus == "" {
			dbus = "unix:path=/run/user/" + strconv.Itoa(uid) + "/bus"
		}
		cmd.Env = append(os.Environ(),
			"DISPLAY="+display,
			"DBUS_SESSION_BUS_ADDRESS="+dbus,
		)
		cmd.Run()
	}
}
