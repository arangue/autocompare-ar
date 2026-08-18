package domain

import "time"

type Listing struct {
	ID         int64
	Source     string
	ExternalID string
	TrimID     *int
	Year       int
	KM         *int
	Price      int64
	Currency   string
	Location   *string
	URL        *string
	LastSeenAt time.Time
}

type MarketSummary struct {
	Count   int
	Median  *int64
	Minimum *int64
	Maximum *int64
	P25     *int64
	P75     *int64
}
