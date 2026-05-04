<div align="center">

# 🎌 ani-rem

**A lightweight anime airing reminder for Linux (cli) — built in Go. (dashboard version coming soon)**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Linux-FCC624?style=flat-square&logo=linux&logoColor=black)](https://www.linux.org/)
[![API](https://img.shields.io/badge/Powered%20by-Jikan%20API-E85D8A?style=flat-square)](https://jikan.moe/)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)

Never miss an episode again. `ani-rem` runs silently in the background and fires a persistent desktop notification when your next episode is getting close.

</div>

---

## Table of Contents

- [Features](#features)
- [How It Works](#how-it-works)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration & Storage](#configuration--storage)
- [Adding to Startup Applications](#adding-to-startup-applications-linux-mint)
- [Project Structure](#project-structure)
- [Credits](#credits)

---

## Features

- 🔍 **Interactive Search & Add** — Search anime by name via the Jikan API with a Promptui-driven interactive menu. Optionally view the synopsis before adding.
- 📋 **Watchlist Management** — List all saved anime, view details, delete individual entries, or wipe the entire list.
- ⏱️ **Countdown Timers** — Displays time-until-next-episode countdowns, but only for shows with status `"Currently Airing"`. Finished or upcoming shows show their status instead.
- 🔔 **Background Notifications** — A persistent daemon checks your watchlist every 5 minutes and sends a `critical`-priority desktop notification for any episode airing within the next 24 hours.
- 🔕 **Notification Deduplication** — A lock-file system prevents duplicate alerts. Once a notification is sent for a show, it won't fire again for at least 1 hour.
- 🕐 **JST → Local Time Conversion** — Converts Japanese Standard Time broadcast schedules to accurate local countdowns.
- 🛑 **Stop Command** — Kill the background daemon cleanly at any time.
- 🖥️ **Interactive Main Menu** — Running `ani-rem` with no subcommand drops you into a Promptui main menu for all actions.

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

The parent process re-launches itself with `ANI_REM_CHILD=1` in its environment. The child runs the worker loop detached from your terminal. `ani-rem stop` reads `/tmp/ani-rem.pid` and sends `kill <PID>` to shut it down.

### Notification Threshold

The worker calls `CheckAiringAnime()` every 5 minutes. For each `"Currently Airing"` show, it parses the countdown returned by `GetTimeUntilAiring()` into a `time.Duration` and fires a notification if the episode airs **within the next 24 hours**:

```
remaining < 24h  →  Send notification: "Critical update: <title> anime releasing soon in <duration>"
```

### Notification Deduplication

To avoid spamming you every 5 minutes, `ani-rem` uses a lock-file per show stored in `/tmp`:

| File | Purpose |
|---|---|
| `/tmp/notify_<Show_Name>.lock` | Timestamp of the last sent notification |

Before sending, `ShouldSendNotification()` checks whether the lock file is older than **1 hour**. If it was sent recently, the notification is skipped and a log line is printed instead. After a successful send, `MarkAsSent()` touches the lock file to reset the timer.

### Notifications from a Background Process

Since the daemon runs outside any interactive session, `notify-send` needs to know which display and D-Bus session to target. `ani-rem` injects these explicitly:

```go
cmd.Env = append(os.Environ(),
    "DISPLAY=:0",
    "DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus",
)
```

The `-u critical` flag makes notifications **persistent** — they stay on screen until dismissed.

### JST Time Conversion

Jikan returns broadcast data like `day: "Saturdays"` and `time: "23:00"`. `ani-rem` converts this to a local countdown by:

1. Loading the `Asia/Tokyo` timezone with `time.LoadLocation`.
2. Mapping the weekday string (e.g. `"Saturdays"` → `time.Saturday`) via a `map[string]time.Weekday`.
3. Calculating the number of days until the next occurrence of that weekday and time.
4. If that time has already passed today, adding 7 days to get next week's airing.
5. Returning the difference formatted as `"Next episode in Xh Ym"`.

---

## Prerequisites

| Requirement | Purpose | Install |
|---|---|---|
| **Go 1.21+** | Build & install the tool | [go.dev/dl](https://go.dev/dl/) |
| **libnotify-bin** | Provides `notify-send` for desktop alerts | See below |

**Install `libnotify-bin`:**

```bash
sudo apt update && sudo apt install libnotify-bin
```

---

## Installation

```bash
go install github.com/sarwan/ani-rem@latest
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

Launches the Promptui main menu with all options: Search & Add Anime, View My Watchlist, Start Background Worker, Stop Background Worker, Exit.

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

Spawns a detached background worker that checks your watchlist every **5 minutes** and fires a desktop notification for any currently-airing show with an episode dropping within the next **24 hours**. Notifications are deduplicated — each show won't alert more than once per hour. Safe to run from a terminal, `.desktop` file, or startup script.

---

### `stop` — Kill the background daemon

```bash
ani-rem stop
```

Reads `/tmp/ani-rem.pid` and kills the running worker process. If no PID file is found, it reports that no active worker was found.

---

### `check` — Run a one-off airing check

```bash
ani-rem check
```

Manually triggers a single pass of `CheckAiringAnime()` — the same logic the background daemon runs. Prints countdown info to the terminal and sends notifications for any currently-airing show within the next 24 hours (subject to the 1-hour deduplication cooldown).

---

## Configuration & Storage

`ani-rem` stores everything locally — no account, no cloud, no telemetry.

| Path | Contents |
|---|---|
| `~/.config/ani-rem/list.json` | Your saved watchlist |
| `/tmp/ani-rem.pid` | PID of the running background daemon (auto-created on start, deleted on stop) |
| `/tmp/notify_<Show_Name>.lock` | Per-show notification cooldown tracker (auto-managed) |

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

## Adding to Startup Applications (Linux Mint)

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

3. Click **Add**, then **Close**.

> Replace `<your-username>` with your actual username. Run `which ani-rem` to confirm the full path to the binary.

---

## Project Structure

```
ani-rem/
├── main.go               # Entry point — calls cmd.Execute()
├── cmd/
│   ├── root.go           # Root command (interactive menu) + start & stop logic
│   ├── create.go         # `create` subcommand — search & add anime
│   ├── list.go           # `list` subcommand — view, detail, delete
│   └── check.go          # `check` subcommand — manual one-off airing check
├── utils/
│   ├── search_anime.go        # Jikan API client
│   ├── storage.go             # JSON read/write for ~/.config/ani-rem/list.json
│   ├── time.go                # JST broadcast string → local countdown
│   ├── notify.go              # notify-send wrapper with deduplication logic
│   └── CheckAiringAnime.go    # Core check loop — parses countdowns, triggers notifications
└── models/
    └── models.go         # AnimeData, Broadcast, JikanResponse structs
```

**Dependencies:**

| Package | Role |
|---|---|
| [`github.com/spf13/cobra`](https://github.com/spf13/cobra) | CLI command structure |
| [`github.com/manifoldco/promptui`](https://github.com/manifoldco/promptui) | Interactive terminal menus |
| Jikan API v4 | Anime metadata (no API key required) |
| `notify-send` (system binary) | Desktop notification delivery |

---

## Credits

Built by **Sarwan Azhar** - a 17-year-old full-stack developer.

Anime data provided by the [Jikan API](https://jikan.moe/), an unofficial MyAnimeList API. Notifications via [`libnotify`](https://gitlab.gnome.org/GNOME/libnotify).