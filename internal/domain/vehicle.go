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

type TrimSearchResult struct {
	TrimID         int    `json:"trim_id"`
	TrimName       string `json:"trim_name"`
	BrandName      string `json:"brand_name"`
	ModelName      string `json:"model_name"`
	GenerationName string `json:"generation_name"`
	YearFrom       *int   `json:"year_from"`
	YearTo         *int   `json:"year_to"`
}
