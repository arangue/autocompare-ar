package domain

import "time"

type Brand struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Model struct {
	ID        int       `json:"id"`
	BrandID   int       `json:"brand_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Trim struct {
	ID           int       `json:"id"`
	GenerationID int       `json:"generation_id"`
	Name         string    `json:"name"`
	YearFrom     *int      `json:"year_from"`
	YearTo       *int      `json:"year_to"`
	CreatedAt    time.Time `json:"created_at"`
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

type TrimDetail struct {
	TrimID         int    `json:"trim_id"`
	TrimName       string `json:"trim_name"`
	BrandID        int    `json:"brand_id"`
	BrandName      string `json:"brand_name"`
	ModelID        int    `json:"model_id"`
	ModelName      string `json:"model_name"`
	GenerationID   int    `json:"generation_id"`
	GenerationName string `json:"generation_name"`
	YearFrom       *int   `json:"year_from"`
	YearTo         *int   `json:"year_to"`
	RequestedYear  *int   `json:"requested_year"`
}
