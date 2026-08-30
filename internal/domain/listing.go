package domain

import "time"

type Listing struct {
	ID         int64     `json:"id"`
	Source     string    `json:"source"`
	ExternalID string    `json:"external_id"`
	TrimID     int       `json:"trim_id"`
	Year       int       `json:"year"`
	KM         *int      `json:"km"`
	Price      int64     `json:"price"`
	Currency   string    `json:"currency"`
	Location   *string   `json:"location"`
	URL        *string   `json:"url"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

type MarketSummary struct {
	Count   int
	Median  *int64
	Minimum *int64
	Maximum *int64
	P25     *int64
	P75     *int64
}
