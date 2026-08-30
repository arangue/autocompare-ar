package domain

import "time"

type ReferencePrice struct {
	ID         int64     `json:"id"`
	TrimID     int       `json:"trim_id"`
	Year       int       `json:"year"`
	Source     string    `json:"source"`
	Kind       string    `json:"kind"`
	Price      int64     `json:"price"`
	Currency   string    `json:"currency"`
	ObservedAt time.Time `json:"observed_at"`
}
