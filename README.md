<div align="center">

# 🎌 ani-rem

**A lightweight anime airing reminder for Linux, macOS & Windows (CLI) — built in Go.**

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-2E8B57?style=flat-square&logo=linux&logoColor=white)](https://www.linux.org/)
[![API](https://img.shields.io/badge/Powered%20by-Jikan%20API-E85D8A?style=flat-square)](https://jikan.moe/)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)

Never miss an episode again. `ani-rem` runs silently in the background, fires persistent desktop notifications when your next episode is getting close, and can sync your airing schedule directly to **Google Calendar**.

</div>

---

## Table of Contents

- [Features](#features)
- [How It Works](#how-it-works)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
- [Google Calendar Integration](#google-calendar-integration)
- [Configuration & Storage](#configuration--storage)
- [Adding to Startup Applications](#adding-to-startup-applications)
- [Project Structure](#project-structure)
- [Credits](#credits)

---

## Features

- 🔍 **Interactive Search & Add** — Search anime by name via the Jikan API with a Promptui-driven interactive menu. Optionally view the synopsis before adding.
- 📅 **Seasonal Anime Browser** — Browse and bulk-add anime from the current or any specific season (winter, spring, summer, fall). Multi-select interface with filtering options.
- 📋 **Watchlist Management** — List all saved anime, view details, delete individual entries, or wipe the entire list.
- ⏱️ **Countdown Timers** — Displays time-until-next-episode countdowns, but only for shows with status `"Currently Airing"`. Finished or upcoming shows show their status instead.
- 🔔 **Background Notifications** — A persistent daemon checks your watchlist every 5 minutes and sends a desktop notification for any currently-airing show with an episode dropping within the configured threshold (default: 24 hours).
- 🔕 **Notification Deduplication** — A lock-file system prevents duplicate alerts. Once a notification is sent for a show, it won't fire again for at least 1 hour.
- 🕐 **JST → Local Time Conversion** — Converts Japanese Standard Time broadcast schedules to accurate local countdowns.
- 📅 **Google Calendar Sync** — One-way sync your currently airing anime as recurring weekly events to any Google Calendar. Supports auto-sync from the background worker.
- 🗑️ **Calendar Cleanup** — Remove specific anime events or clear all ani-rem events from your calendar in one go.
- ⚙️ **Configurable Settings** — Change notification threshold, toggle auto-sync, and manage preferences via an interactive menu or your `$EDITOR`.
- 🛑 **Stop Command** — Kill the background daemon cleanly at any time.
- 🖥️ **Interactive Main Menu** — Running `ani-rem` with no subcommand drops you into a Promptui main menu for all actions.
- 🖥️ **Cross-Platform Support** — Works on **Linux**, **macOS**, and **Windows** with native desktop notifications on each platform.

---

## How It Works

### The Double-Start Daemon

`ani-rem start` uses a **"Double-Start"** pattern to spawn a true background process that survives after you close your terminal.

```
ani-rem start
    │
    ├─► Is ANI_REM_CHILD=1 set? ──NO──► Re-launch self with ANI_REM_CHILD=1
    │                                    Write child PID → /tmp/ani-rem.pid
    │                                    Print success, parent exits.
    │
    └─► YES ─► Enter the worker loop (checks every 5 minutes)
```

The parent process re-launches itself with `ANI_REM_CHILD=1` in its environment. The child runs the worker loop detached from your terminal. `ani-rem stop` reads the PID file and sends the appropriate signal to shut it down.

### Notification Threshold

The worker calls `CheckAiringAnime()` every 5 minutes. For each `"Currently Airing"` show, it parses the countdown returned by `GetTimeUntilAiring()` into a `time.Duration` and fires a notification if the episode airs within the configured threshold (default **24 hours**, customizable via `ani-rem config`):

```
remaining < threshold  →  Send notification: "ani-rem: <title> — Episode releasing soon in <duration>"
```

### Notification Deduplication

To avoid spamming you every 5 minutes, `ani-rem` uses a lock-file per show stored in the OS temp directory:

| File | Purpose |
|---|---|
| `{temp}/notify_<Show_Name>.lock` | Timestamp of the last sent notification |

Before sending, `ShouldSendNotification()` checks whether the lock file is older than **1 hour**. If it was sent recently, the notification is skipped and a log line is printed instead. After a successful send, `MarkAsSent()` touches the lock file to reset the timer.

### Cross-Platform Notifications

`ani-rem` uses [`github.com/gen2brain/beeep`](https://github.com/gen2brain/beeep) for cross-platform desktop notifications, with platform-specific fallbacks:

| Platform | Primary Method | Fallback |
|---|---|---|
| **Linux** | `beeep.Notify()` | `notify-send` with explicit `DISPLAY` and `DBUS_SESSION_BUS_ADDRESS` |
| **macOS** | `beeep.Notify()` | `osascript` display notification |
| **Windows** | `beeep.Notify()` | PowerShell `Windows.UI.Notifications` toast |

On Linux, the `-u critical` flag makes notifications **persistent** — they stay on screen until dismissed.

### JST Time Conversion

Jikan returns broadcast data like `day: "Saturdays"` and `time: "23:00"`. `ani-rem` converts this to a local countdown by:

1. Loading the `Asia/Tokyo` timezone with `time.LoadLocation`.
2. Mapping the weekday string (e.g. `"Saturdays"` → `time.Saturday`) via a `map[string]time.Weekday`.
3. Calculating the number of days until the next occurrence of that weekday and time.
4. If that time has already passed today, adding 7 days to get next week's airing.
5. Returning the difference formatted as `"Next episode in Xh Ym"`.

### Google Calendar Sync

See the [Google Calendar Integration](#google-calendar-integration) section for full details. In summary:

When you sync an anime, `ani-rem` creates a **recurring weekly event** in your chosen Google Calendar:

- **Title:** `📺 <Anime Title> - New Episode`
- **Duration:** 1 hour
- **Recurrence:** Weekly for a configurable number of weeks (default: 12)
- **Description:** Includes airing time, status, MAL score, truncated synopsis, and MAL ID
- **Location:** `Online Streaming (Crunchyroll, Funimation, Netflix, etc.)`
- **Duplicate Prevention:** Before syncing, `ani-rem` searches existing events and skips if the anime is already present. To update an existing schedule (e.g., change the number of weeks), remove it first with `ani-rem calendar remove`, then re-sync.

The background worker supports **auto-sync** (`--auto-sync` flag or config toggle) which runs once per day, keeping your calendar up-to-date with your watchlist automatically.

---

## Prerequisites

| Requirement | Purpose | Install |
|---|---|---|
| **Go 1.25+** | Build & install the tool | [go.dev/dl](https://go.dev/dl/) |
| **libnotify-bin** *(Linux only)* | Provides `notify-send` for desktop alerts fallback | See below |

**Install `libnotify-bin` (Linux):**

```bash
sudo apt update && sudo apt install libnotify-bin
```

---

## Installation

```bash
# Clone the repository
git clone https://github.com/sarwanazhar/ani-rem.git
cd ani-rem

# Build and install
go build && go install
```

This places the binary in `~/go/bin/`. Make sure that directory is in your `$PATH`:

```bash
# Add to your ~/.bashrc or ~/.zshrc if not already present
export PATH="$PATH:$HOME/go/bin"
```

**Verify the install:**

```bash
ani-rem --help
```

---

## Usage

### Interactive Menu (no subcommand)

```bash
ani-rem
```

Launches the Promptui main menu with all options: Search & Add Anime, Browse Seasonal Anime, View My Watchlist, Start Background Worker, Stop Background Worker, Google Calendar, Settings, Exit.

---

### `create` — Search and add an anime

```bash
ani-rem create
# or pass the name directly with a flag:
ani-rem create -n "Dandadan"
```

Opens an interactive search powered by the Jikan API. After selecting a result you get an action menu:

- **Confirm & Add to List** — saves it to `~/.config/ani-rem/list.json`
- **Show Details** — prints the status, score, and synopsis, then asks if you want to add it
- **Do Nothing** — cancels without saving

---

### `seasonal` — Browse seasonal anime

```bash
ani-rem seasonal                    # Browse current season
ani-rem seasonal --year 2024 --season spring
ani-rem seasonal --filter           # Show only currently airing
ani-rem seasonal --page 2           # Navigate pages
```

Browse anime from any season with an interactive multi-select interface:

- Navigate with arrow keys, press **Enter** to select/deselect anime
- View synopsis and details for each show
- Bulk add selected anime to your watchlist
- Filter to show only currently airing shows

**Interactive controls:**
- `☑/☐` — Toggle selection
- `✅ Add Selected to Watchlist` — Confirm and add all selected
- `🔄 Select All / Deselect All` — Toggle all items at once

---

### `seasonal bulk` — Quick bulk-add seasonal anime

```bash
ani-rem seasonal bulk                    # Interactive multi-select
ani-rem seasonal bulk --all              # Add all currently airing
ani-rem seasonal bulk --min-score 7.5    # Filter by minimum score
ani-rem seasonal bulk --yes              # Skip confirmation prompt
```

A streamlined interface for quickly adding multiple current-season shows:

- `--all` / `-a` — Automatically select all currently airing anime
- `--min-score` / `-m` — Filter by minimum MAL score (0-10)
- `--yes` / `-y` — Skip the confirmation prompt
- `--filter` / `-f` — Only show currently airing anime

---

### `list` — View and manage your watchlist

```bash
ani-rem list
```

Displays all saved anime in an interactive Promptui menu. Selecting a show gives you:

- **Show Details** — prints the status, countdown (only shown if `"Currently Airing"`), and synopsis
- **Delete from List** — removes that entry from the JSON file
- **Back** — returns to the watchlist

From the list you can also select **🗑️ Delete Entire List** to remove the `list.json` file entirely, or **➜ Exit to Menu** to return.

**Example details output:**
```
--- Dandadan ---
Status: Currently Airing
Next Airing: Next episode in 4h 22m

Synopsis: ...
```

---

### `start` — Launch the background daemon

```bash
ani-rem start
```

Spawns a detached background worker that checks your watchlist every **5 minutes** and fires a desktop notification for any currently-airing show with an episode dropping within the configured notification threshold (default: **24 hours**). Notifications are deduplicated — each show won't alert more than once per hour. Safe to run from a terminal, startup script, or OS startup application settings.

**With auto-calendar-sync:**
```bash
ani-rem start --auto-sync
```
This will also sync your currently airing anime to Google Calendar **once per day** automatically (requires calendar to be connected first).

---

### `stop` — Kill the background daemon

```bash
ani-rem stop
```

Reads the PID file and terminates the running worker process. If no PID file is found, it reports that no active worker was found.

---

### `check` — Run a one-off airing check

```bash
ani-rem check
```

Manually triggers a single pass of `CheckAiringAnime()` — the same logic the background daemon runs. Prints countdown info to the terminal and sends notifications for any currently-airing show within the notification threshold (subject to the 1-hour deduplication cooldown).

---

### `config` — Manage settings

```bash
ani-rem config
```

Opens an interactive settings menu where you can:

| Option | Description |
|---|---|
| 📄 View current settings | Display the current configuration as JSON |
| ✏️ Edit settings in `$EDITOR` | Open `~/.config/ani-rem/config.json` in your default editor |
| 🔄 Auto-sync toggle | Enable/disable daily auto-sync to Google Calendar |
| ⏰ Change notification threshold | Set how many hours before airing to trigger notifications (default: 24) |

**Configuration file** (`~/.config/ani-rem/config.json`):
```json
{
  "auto_sync": false,
  "calendar_id": "",
  "notification_threshold_hours": 24
}
```

---

## Google Calendar Integration

`ani-rem` can sync your currently airing anime schedule to Google Calendar as recurring weekly events. This requires a one-time OAuth 2.0 setup.

### Quick Setup

```bash
# Step 1: View the guided setup instructions
ani-rem setup-calendar

# Step 2: Connect your Google account
ani-rem calendar connect
# (You will be prompted for Client ID and Secret, then a browser opens for OAuth)

# Step 3: Sync your airing anime
ani-rem sync
```

### `setup-calendar` — Guided OAuth Setup

```bash
ani-rem setup-calendar
```

Prints step-by-step instructions for creating a Google Cloud Project, enabling the Google Calendar API, configuring the OAuth consent screen, and generating Desktop App credentials. No API calls are made — this is purely informational.

---

### `calendar` — Calendar Management Menu

```bash
ani-rem calendar
```

Opens an interactive menu with the following options:

| Option | Description |
|---|---|
| 🔐 Connect / Sign in (`calendar connect`) | Authenticate with Google OAuth |
| 📅 List my calendars | Show all calendars in your account |
| 🔄 Sync anime to calendar (`sync`) | Jump to the `sync` command |
| 🗑️ Clear all anime events (`calendar clear`) | Delete **all** events created by ani-rem |
| ❌ Remove specific anime events (`calendar remove`) | Pick one anime and delete only its events |
| 🚫 Disconnect (`calendar disconnect`) | Remove stored OAuth tokens (events stay in Google Calendar) |

---

### `calendar connect` — Authenticate

```bash
ani-rem calendar connect
```

Prompts for your Google OAuth Client ID and Client Secret, saves them to `~/.config/ani-rem/google_credentials.json`, then launches a local HTTP server on `localhost:8080` to handle the OAuth callback. A browser window opens automatically. After granting permission, your token is saved to `~/.config/ani-rem/google_token.json`.

---

### `calendar disconnect` — Remove Access

```bash
ani-rem calendar disconnect
```

Deletes the locally stored OAuth token. Your anime events remain in Google Calendar, but `ani-rem` will no longer be able to add or remove events until you re-authenticate.

---

### `sync` — Sync anime to Google Calendar

```bash
ani-rem sync              # Interactive mode — pick anime from your watchlist
ani-rem sync --all        # Sync all currently airing anime
ani-rem sync --anime "One Piece"   # Sync a specific anime by exact title
ani-rem sync --weeks 24   # Schedule for the next 24 weeks (default: 12)
ani-rem sync --calendar "your.calendar.id@gmail.com"  # Target a specific calendar
```

**Interactive mode** presents a menu of all your `"Currently Airing"` anime, plus an option to **Sync All**. After selection, you must confirm before events are created.

**Duplicate prevention:** `ani-rem` searches your calendar for existing events with the same title before creating new ones. If found, the anime is skipped. To update an existing schedule (e.g., change the number of weeks), remove it first with `ani-rem calendar remove`, then re-sync.

---

### `calendar clear` — Delete all ani-rem events

```bash
ani-rem calendar clear
# or force without confirmation:
ani-rem calendar clear --force
```

Searches your primary calendar for all events whose summary contains `📺` and `- New Episode` (or whose description contains `Powered by ani-rem`), lists them, and asks for confirmation before deleting. Use `--force` / `-f` to skip the confirmation prompt.

---

### `calendar remove` — Delete events for a specific anime

```bash
ani-rem calendar remove
```

Shows a list of your currently airing anime from your local watchlist. Pick one, confirm, and all recurring events for that title will be removed from your primary calendar.

---

## Configuration & Storage

`ani-rem` stores everything locally — no account, no cloud, no telemetry.

| Path | Contents | Permissions |
|---|---|---|
| `~/.config/ani-rem/list.json` | Your saved watchlist | `0644` |
| `~/.config/ani-rem/config.json` | App settings (auto-sync, threshold, calendar ID) | `0644` |
| `~/.config/ani-rem/google_credentials.json` | Google OAuth Client ID & Secret | `0600` |
| `~/.config/ani-rem/google_token.json` | Google OAuth access & refresh tokens | `0600` |
| `{temp}/ani-rem.pid` | PID of the running background daemon (auto-created on start, deleted on stop) | `0600` |
| `{temp}/ani-rem-last-sync` | Timestamp of the last auto-sync (for daily deduplication) | `0644` |
| `{temp}/notify_<Show_Name>.lock` | Per-show notification cooldown tracker (auto-managed) | `0644` |

> **Security note:** `google_credentials.json` and `google_token.json` contain sensitive OAuth credentials and are stored with `0600` permissions (owner read/write only). Do not share or commit these files.

**Example `list.json` entry:**

```json
[
  {
    "mal_id": 53887,
    "title": "Dandadan",
    "status": "Currently Airing",
    "airing": true,
    "broadcast": {
      "day": "Saturdays",
      "time": "23:00",
      "timezone": "Asia/Tokyo",
      "string": "Saturdays at 23:00 (JST)"
    },
    "score": 8.7,
    "synopsis": "..."
  }
]
```

Duplicate entries are prevented automatically — if you try to add an anime whose `mal_id` is already saved, `ani-rem` warns you and skips it.

---

## Adding to Startup Applications

### Linux Mint / GNOME

To have the daemon start automatically on login:

1. Open **Menu → Preferences → Startup Applications**, or run:
   ```bash
   gnome-session-properties
   ```

2. Click **Add** and fill in:

   | Field | Value |
   |---|---|
   | **Name** | `ani-rem` |
   | **Command** | `/home/<your-username>/go/bin/ani-rem start` |
   | **Comment** | `Anime airing reminder background worker` |

   *(Add `--auto-sync` to the command if you want daily calendar sync too.)*

3. Click **Add**, then **Close**.

> Replace `<your-username>` with your actual username. Run `which ani-rem` to confirm the full path to the binary.

### macOS

1. Open **System Settings → General → Login Items**.
2. Click **+** and add the `ani-rem` binary path.
3. Alternatively, create a LaunchAgent plist at `~/Library/LaunchAgents/com.ani-rem.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.ani-rem</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/ani-rem</string>
        <string>start</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
```

Then run:
```bash
launchctl load ~/Library/LaunchAgents/com.ani-rem.plist
```

### Windows

1. Press `Win + R`, type `shell:startup`, and press Enter.
2. Create a shortcut to `ani-rem.exe` with the `start` argument:
   ```
   C:\path	oni-rem.exe start
   ```
3. Alternatively, use Task Scheduler to run `ani-rem start` at logon.

---

## Project Structure

```
ani-rem/
├── main.go                    # Entry point — calls cmd.Execute()
├── cmd/
│   ├── root.go                # Root command (interactive menu) + start, stop & auto-sync logic
│   ├── create.go              # `create` subcommand — search & add anime
│   ├── list.go                # `list` subcommand — view, detail, delete
│   ├── check.go               # `check` subcommand — manual one-off airing check
│   ├── config.go              # `config` subcommand — settings management
│   ├── calendar.go            # `calendar` subcommand — menu & connect/disconnect
│   ├── setup.go               # `setup-calendar` subcommand — guided OAuth instructions
│   ├── sync.go                # `sync` subcommand — sync airing anime to Google Calendar
│   ├── clear.go               # `calendar clear` subcommand — bulk delete events
│   ├── calendar_remove.go     # `calendar remove` subcommand — delete specific anime events
│   ├── seasonal.go            # `seasonal` subcommand — browse and add seasonal anime
│   └── seasonal_bulk.go       # `seasonal bulk` subcommand — quick bulk-add current season
├── utils/
│   ├── search_anime.go        # Jikan API client for search
│   ├── seasonal_anime.go      # Jikan seasonal API client & bulk add logic
│   ├── storage.go             # JSON read/write for ~/.config/ani-rem/list.json
│   ├── config.go              # Configuration file handling (load/save defaults)
│   ├── time.go                # JST broadcast string → local countdown
│   ├── notify.go              # Cross-platform desktop notifications with deduplication
│   ├── check_airing_anime.go  # Core check loop — parses countdowns, triggers notifications
│   └── google_calendar.go     # Google Calendar API client, OAuth flow, event CRUD
└── models/
    └── models.go              # AnimeData, Broadcast, JikanResponse, SeasonalResponse, SeasonListItem structs
```

**Dependencies:**

| Package | Role |
|---|---|
| [`github.com/spf13/cobra`](https://github.com/spf13/cobra) | CLI command structure |
| [`github.com/manifoldco/promptui`](https://github.com/manifoldco/promptui) | Interactive terminal menus |
| [`golang.org/x/oauth2`](https://pkg.go.dev/golang.org/x/oauth2) | Google OAuth 2.0 flow |
| [`google.golang.org/api/calendar/v3`](https://pkg.go.dev/google.golang.org/api/calendar/v3) | Google Calendar API client |
| [`github.com/gen2brain/beeep`](https://github.com/gen2brain/beeep) | Cross-platform desktop notifications |
| Jikan API v4 | Anime metadata (no API key required) |
| `notify-send` *(Linux fallback)* | Desktop notification delivery |

---

## Credits

Built by **Sarwan Azhar** — a 17-year-old full-stack developer.

Anime data provided by the [Jikan API](https://jikan.moe/), an unofficial MyAnimeList API. Cross-platform notifications via [`beeep`](https://github.com/gen2brain/beeep). Calendar integration powered by the [Google Calendar API](https://developers.google.com/calendar).