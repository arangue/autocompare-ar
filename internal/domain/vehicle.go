package domain

import "time"

type Brand struct {
	ID        int
	Name      string
	CreatedAt time.Time
}

type Model struct {
	ID        int
	BrandID   int
	Name      string
	CreatedAt time.Time
}

type Trim struct {
	ID           int
	GenerationID int
	Name         string
	YearFrom     *int
	YearTo       *int
	CreatedAt    time.Time
}
