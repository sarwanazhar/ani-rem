package utils

import (
	"ani-rem/models"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func SearchAnime(query string) ([]models.AnimeData, error) {
	apiURL := fmt.Sprintf("https://api.jikan.moe/v4/anime?q=%s", url.QueryEscape(query))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jikan API returned %d: %s", resp.StatusCode, resp.Status)
	}

	var result models.JikanResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	return result.Data, nil
}
