package utils

import (
	"ani-rem/models"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func SearchAnime(query string) ([]models.AnimeData, error) {
	apiURL := fmt.Sprintf("https://api.jikan.moe/v4/anime?q=%s", url.QueryEscape(query))

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result models.JikanResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data, nil
}
