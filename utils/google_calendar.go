package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"ani-rem/models"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const (
	calendarScope = calendar.CalendarScope
	redirectURI   = "http://localhost:8080/oauth2callback"
)

type GoogleCalendarClient struct {
	service *calendar.Service
	ctx     context.Context
	config  *oauth2.Config
	token   *oauth2.Token
}

type TokenStore struct {
	path string
}

func NewTokenStore() *TokenStore {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config", "ani-rem")
	os.MkdirAll(configDir, 0755)
	return &TokenStore{path: filepath.Join(configDir, "google_token.json")}
}

func (ts *TokenStore) Save(token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return os.WriteFile(ts.path, data, 0600)
}

func (ts *TokenStore) Load() (*oauth2.Token, error) {
	data, err := os.ReadFile(ts.path)
	if err != nil {
		return nil, err
	}
	var token oauth2.Token
	err = json.Unmarshal(data, &token)
	return &token, err
}

func (ts *TokenStore) Delete() error {
	if _, err := os.Stat(ts.path); err == nil {
		return os.Remove(ts.path)
	}
	return nil
}

type CredentialsStore struct {
	path string
}

func NewCredentialsStore() *CredentialsStore {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config", "ani-rem")
	os.MkdirAll(configDir, 0755)
	return &CredentialsStore{path: filepath.Join(configDir, "google_credentials.json")}
}

func (cs *CredentialsStore) Save(clientID, clientSecret string) error {
	data := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(cs.path, jsonData, 0600)
}

func (cs *CredentialsStore) Load() (clientID, clientSecret string, err error) {
	data, err := os.ReadFile(cs.path)
	if err != nil {
		return "", "", err
	}
	var creds map[string]string
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", "", err
	}
	return creds["client_id"], creds["client_secret"], nil
}

func (cs *CredentialsStore) Delete() error {
	if _, err := os.Stat(cs.path); err == nil {
		return os.Remove(cs.path)
	}
	return nil
}

func NewGoogleCalendarClient() (*GoogleCalendarClient, error) {
	ctx := context.Background()

	credsStore := NewCredentialsStore()
	clientID, clientSecret, err := credsStore.Load()
	if err != nil {
		return nil, fmt.Errorf("no credentials found. Please run 'ani-rem calendar connect' first")
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  redirectURI,
		Scopes:       []string{calendarScope},
	}

	tokenStore := NewTokenStore()
	token, err := tokenStore.Load()

	client := &GoogleCalendarClient{
		ctx:    ctx,
		config: config,
		token:  token,
	}

	if err == nil && token != nil && token.Valid() {
		httpClient := config.Client(ctx, token)
		service, err := calendar.NewService(ctx, option.WithHTTPClient(httpClient))
		if err != nil {
			return nil, err
		}
		client.service = service
		return client, nil
	}

	return client, nil
}

func NewEmptyClient() *GoogleCalendarClient {
	return &GoogleCalendarClient{
		ctx: context.Background(),
	}
}

func (c *GoogleCalendarClient) SetCredentials(clientID, clientSecret string) error {
	if c == nil {
		return fmt.Errorf("client is nil")
	}
	credsStore := NewCredentialsStore()
	if err := credsStore.Save(clientID, clientSecret); err != nil {
		return err
	}
	c.config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  redirectURI,
		Scopes:       []string{calendarScope},
	}
	return nil
}

func (c *GoogleCalendarClient) IsAuthenticated() bool {
	return c != nil && c.service != nil && c.token != nil && c.token.Valid()
}

func (c *GoogleCalendarClient) Authenticate() error {
	if c == nil {
		return fmt.Errorf("client is nil, call NewEmptyClient() first")
	}
	if c.IsAuthenticated() {
		fmt.Println("✓ Already authenticated!")
		return nil
	}
	if c.config == nil {
		return fmt.Errorf("no config set; run SetCredentials first")
	}

	authChan := make(chan string)
	errChan := make(chan error)

	mux := http.NewServeMux()
	server := &http.Server{Addr: ":8080", Handler: mux}

	mux.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, "<html><body><h2>✅ Authentication successful!</h2><p>You can close this window and return to the terminal.</p></body></html>")
			authChan <- code
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			go server.Shutdown(ctx)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "<html><body><h2>❌ Authentication failed</h2><p>No authorization code received.</p></body></html>")
			errChan <- fmt.Errorf("no code in callback")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			go server.Shutdown(ctx)
		}
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	time.Sleep(100 * time.Millisecond)

	authURL := c.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	fmt.Println("\n🔐 Google Calendar Authentication")
	fmt.Println("=================================")
	fmt.Printf("1. Opening browser for you to sign in...\n")
	fmt.Printf("2. If browser doesn't open, visit:\n%s\n", authURL)
	fmt.Println("3. Grant access to ani-rem")
	fmt.Println("\n🌐 Opening browser...")

	openBrowser(authURL)

	select {
	case code := <-authChan:
		token, err := c.config.Exchange(context.Background(), code)
		if err != nil {
			return fmt.Errorf("token exchange failed: %v", err)
		}
		c.token = token
		if err := NewTokenStore().Save(token); err != nil {
			return fmt.Errorf("failed to save token: %v", err)
		}
		httpClient := c.config.Client(c.ctx, token)
		service, err := calendar.NewService(c.ctx, option.WithHTTPClient(httpClient))
		if err != nil {
			return err
		}
		c.service = service
		fmt.Println("\n✅ Successfully connected to Google Calendar!")
		return nil
	case err := <-errChan:
		return fmt.Errorf("authentication error: %v", err)
	case <-time.After(5 * time.Minute):
		return fmt.Errorf("authentication timeout (5 minutes)")
	}
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Printf("Please manually open this URL in your browser:\n%s\n", url)
	}
}

func (c *GoogleCalendarClient) AddEventToCalendar(event *calendar.Event, calendarID string) error {
	_, err := c.service.Events.Insert(calendarID, event).Do()
	return err
}

// CreateAnimeEvent – no reminders
func CreateAnimeEvent(anime models.AnimeData, airingTime time.Time) *calendar.Event {
	endTime := airingTime.Add(time.Hour)

	event := &calendar.Event{
		Summary: fmt.Sprintf("📺 %s - New Episode", anime.Title),
		Description: fmt.Sprintf(
			"Episode airing time: %s\nStatus: %s\nScore: %.2f/10\n\nSynopsis:\n%s\n\nSource: MyAnimeList (ID: %d)\nPowered by ani-rem",
			anime.Broadcast.String, anime.Status, anime.Score,
			truncateString(anime.Synopsis, 500), anime.MalID,
		),
		Location: "Online Streaming (Crunchyroll, Funimation, Netflix, etc.)",
		Start: &calendar.EventDateTime{
			DateTime: airingTime.Format(time.RFC3339),
			TimeZone: "Asia/Tokyo",
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
			TimeZone: "Asia/Tokyo",
		},
		Recurrence: []string{"RRULE:FREQ=WEEKLY;COUNT=12"},
	}
	return event
}

// FindAnimeEventsByTitle searches for existing events with the given anime title
// Returns true if at least one event is found (i.e., already synced)
func (c *GoogleCalendarClient) IsAnimeAlreadySynced(calendarID, animeTitle string) (bool, error) {
	// Search for events from next 6 months (to catch existing recurring events)
	now := time.Now()
	timeMin := now.Format(time.RFC3339)
	timeMax := now.AddDate(0, 6, 0).Format(time.RFC3339)

	events, err := c.ListEvents(calendarID, timeMin, timeMax)
	if err != nil {
		return false, err
	}

	// Look for event summary that matches the pattern
	expectedSummary := fmt.Sprintf("📺 %s - New Episode", animeTitle)
	for _, ev := range events {
		if ev.Summary == expectedSummary {
			return true, nil
		}
	}
	return false, nil
}

func (c *GoogleCalendarClient) SyncAnimeToCalendar(anime models.AnimeData, weeks int, calendarID string) error {
	if !c.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Run 'ani-rem calendar connect' first")
	}
	if anime.Status != "Currently Airing" {
		return fmt.Errorf("anime '%s' is not currently airing", anime.Title)
	}

	// Check if already synced
	alreadySynced, err := c.IsAnimeAlreadySynced(calendarID, anime.Title)
	if err != nil {
		return fmt.Errorf("failed to check existing events: %v", err)
	}
	if alreadySynced {
		return fmt.Errorf("already synced (event exists in calendar)")
	}

	airingTimes, err := calculateNextAiringTimes(anime, weeks)
	if err != nil {
		return err
	}
	if len(airingTimes) == 0 {
		return fmt.Errorf("no future airing times found")
	}

	fmt.Printf("\n📅 Syncing %s to Google Calendar...\n", anime.Title)
	event := CreateAnimeEvent(anime, airingTimes[0])
	event.Recurrence = []string{fmt.Sprintf("RRULE:FREQ=WEEKLY;COUNT=%d", weeks)}
	err = c.AddEventToCalendar(event, calendarID)
	if err != nil {
		return fmt.Errorf("failed to create event: %v", err)
	}
	fmt.Printf("  ✓ Created recurring calendar event for %s (next %d episodes)\n", anime.Title, weeks)
	return nil
}

func (c *GoogleCalendarClient) SyncMultipleAnime(animeList []models.AnimeData, weeks int, calendarID string) error {
	success := 0
	fail := 0
	skipped := 0

	for _, anime := range animeList {
		err := c.SyncAnimeToCalendar(anime, weeks, calendarID)
		if err != nil {
			if strings.Contains(err.Error(), "already synced") {
				fmt.Printf("  ⏭️  Skipped %s: %v\n", anime.Title, err)
				skipped++
			} else {
				fmt.Printf("  ❌ Failed to sync %s: %v\n", anime.Title, err)
				fail++
			}
		} else {
			success++
		}
	}
	fmt.Printf("\n📊 Sync Summary: %d succeeded, %d skipped (already in calendar), %d failed\n", success, skipped, fail)
	return nil
}

func calculateNextAiringTimes(anime models.AnimeData, weeks int) ([]time.Time, error) {
	if anime.Broadcast.Time == "" || anime.Broadcast.Day == "" {
		return nil, fmt.Errorf("no broadcast schedule")
	}
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return nil, err
	}
	targetDay, ok := BroadcastDayMap[anime.Broadcast.Day]
	if !ok {
		return nil, fmt.Errorf("invalid broadcast day: %s", anime.Broadcast.Day)
	}
	var hour, min int
	fmt.Sscanf(anime.Broadcast.Time, "%d:%d", &hour, &min)

	now := time.Now().In(loc)
	daysUntil := (int(targetDay) - int(now.Weekday()) + 7) % 7
	firstAiring := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, loc)
	firstAiring = firstAiring.AddDate(0, 0, daysUntil)
	if firstAiring.Before(now) {
		firstAiring = firstAiring.AddDate(0, 0, 7)
	}
	var times []time.Time
	for i := 0; i < weeks; i++ {
		times = append(times, firstAiring.AddDate(0, 0, i*7))
	}
	return times, nil
}

func (c *GoogleCalendarClient) ListCalendarLists() error {
	if !c.IsAuthenticated() {
		return fmt.Errorf("not authenticated")
	}
	list, err := c.service.CalendarList.List().Do()
	if err != nil {
		return err
	}
	fmt.Println("\n📅 Your Google Calendars:")
	fmt.Println("=========================")
	for _, cal := range list.Items {
		primary := ""
		if cal.Primary {
			primary = " (Primary)"
		}
		fmt.Printf("  • %s%s\n    ID: %s\n", cal.Summary, primary, cal.Id)
	}
	return nil
}

func (c *GoogleCalendarClient) GetPrimaryCalendarID() (string, error) {
	list, err := c.service.CalendarList.List().Do()
	if err != nil {
		return "", err
	}
	for _, cal := range list.Items {
		if cal.Primary {
			return cal.Id, nil
		}
	}
	return "", fmt.Errorf("no primary calendar found")
}

// ListEvents returns events in a calendar between timeMin and timeMax
func (c *GoogleCalendarClient) ListEvents(calendarID, timeMin, timeMax string) ([]*calendar.Event, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}
	events, err := c.service.Events.List(calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(timeMin).
		TimeMax(timeMax).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, err
	}
	return events.Items, nil
}

// DeleteEvent removes an event from a calendar
func (c *GoogleCalendarClient) DeleteEvent(calendarID, eventID string) error {
	if !c.IsAuthenticated() {
		return fmt.Errorf("not authenticated")
	}
	return c.service.Events.Delete(calendarID, eventID).Do()
}

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// DeleteAnimeEvents removes all calendar events for a specific anime title
func (c *GoogleCalendarClient) DeleteAnimeEvents(calendarID, animeTitle string) (int, error) {
	expectedSummary := fmt.Sprintf("📺 %s - New Episode", animeTitle)
	now := time.Now()
	timeMin := now.AddDate(-1, 0, 0).Format(time.RFC3339)
	timeMax := now.AddDate(2, 0, 0).Format(time.RFC3339)

	events, err := c.ListEvents(calendarID, timeMin, timeMax)
	if err != nil {
		return 0, err
	}

	var toDelete []*calendar.Event
	for _, ev := range events {
		if ev.Summary == expectedSummary {
			toDelete = append(toDelete, ev)
		}
	}

	if len(toDelete) == 0 {
		return 0, nil
	}

	deleted := 0
	for _, ev := range toDelete {
		if err := c.DeleteEvent(calendarID, ev.Id); err != nil {
			fmt.Printf("  ⚠️ Failed to delete %s: %v\n", ev.Summary, err)
		} else {
			deleted++
			fmt.Printf("  ✓ Deleted: %s\n", ev.Summary)
		}
	}
	return deleted, nil
}
