package models

type Broadcast struct {
	Day      string `json:"day"`      // e.g., "Tuesdays"
	Time     string `json:"time"`     // e.g., "23:00"
	Timezone string `json:"timezone"` // e.g., "Asia/Tokyo"
	String   string `json:"string"`   // e.g., "Tuesdays at 23:00 (JST)"
}

type AnimeData struct {
	MalID     int       `json:"mal_id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"` // "Finished Airing", "Currently Airing", or "Not yet aired"
	Airing    bool      `json:"airing"` // true or false
	Broadcast Broadcast `json:"broadcast"`
	Score     float64   `json:"score"`
	Synopsis  string    `json:"synopsis"`
}

type JikanResponse struct {
	Data []AnimeData `json:"data"`
}
