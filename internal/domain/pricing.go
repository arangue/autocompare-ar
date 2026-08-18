package domain

import "time"

type ReferencePrice struct {
	ID         int64
	TrimID     *int
	Year       int
	Source     string
	Price      int64
	Currency   string
	ObservedAt time.Time
}
