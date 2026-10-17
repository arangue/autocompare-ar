package domain

import "time"

type Listing struct {
	ID         int64     `json:"id"`
	Source     string    `json:"source"`
	ExternalID string    `json:"external_id"`
	TrimID     int       `json:"trim_id"`
	RawTitle   string    `json:"-"`
	Year       int       `json:"year"`
	KM         *int      `json:"km"`
	Price      int64     `json:"price"`
	Currency   string    `json:"currency"`
	Location   *string   `json:"location"`
	URL        *string   `json:"url"`
	Active     *bool     `json:"active"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

type MarketSource struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
}

type MarketSummary struct {
	Count      int
	Median     *int64
	Minimum    *int64
	Maximum    *int64
	P25        *int64
	P75        *int64
	Currency   string
	ObservedAt *time.Time
	Sources    []MarketSource
}

type TrimMarketSummary struct {
	TrimID     int            `json:"trim_id"`
	Year       int            `json:"year"`
	Count      int            `json:"count"`
	Median     *int64         `json:"median"`
	Minimum    *int64         `json:"minimum"`
	Maximum    *int64         `json:"maximum"`
	P25        *int64         `json:"p25"`
	P75        *int64         `json:"p75"`
	Currency   string         `json:"currency"`
	ObservedAt *time.Time     `json:"observed_at"`
	Sources    []MarketSource `json:"sources"`
}

type ModelOffer struct {
	ModelID   int    `json:"model_id"`
	ModelName string `json:"model_name"`
	BrandID   int    `json:"brand_id"`
	BrandName string `json:"brand_name"`
	Count     int    `json:"count"`
}

type ModelOfferList struct {
	ObservedAt *time.Time   `json:"observed_at"`
	Models     []ModelOffer `json:"models"`
}

type ModelOfferYear struct {
	Year  int `json:"year"`
	Count int `json:"count"`
}

type ModelOfferTrim struct {
	TrimID         int              `json:"trim_id"`
	TrimName       string           `json:"trim_name"`
	GenerationName string           `json:"generation_name"`
	Count          int              `json:"count"`
	Years          []ModelOfferYear `json:"years"`
}

type ModelOfferDetail struct {
	ModelID    int              `json:"model_id"`
	ModelName  string           `json:"model_name"`
	BrandID    int              `json:"brand_id"`
	BrandName  string           `json:"brand_name"`
	Count      int              `json:"count"`
	ObservedAt *time.Time       `json:"observed_at"`
	Trims      []ModelOfferTrim `json:"trims"`
}
