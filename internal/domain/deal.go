package domain

const (
	DealMinSampleSize = 5

	DealStatusUnknown          = "unknown"
	DealStatusUnavailable      = "unavailable"
	DealStatusOK               = "ok"
	DealStatusInsufficientData = "insufficient_data"

	DealBandBelow            = "below"
	DealBandNear             = "near"
	DealBandAbove            = "above"
	DealBandInsufficientData = "insufficient_data"

	DealDisclaimerOK               = "Estimación a partir de publicaciones comparables; el precio publicado no es precio de venta."
	DealDisclaimerInsufficientData = "No hay publicaciones suficientes para estimar el mercado de esta versión/año."
)

type DealMarketSnapshot struct {
	Count  int    `json:"count"`
	Median *int64 `json:"median"`
	P25    *int64 `json:"p25"`
	P75    *int64 `json:"p75"`
}

type DealAssessment struct {
	TrimID     int                `json:"trim_id"`
	Year       int                `json:"year"`
	Price      int64              `json:"price"`
	Currency   string             `json:"currency"`
	Status     string             `json:"status"`
	Band       string             `json:"band"`
	DeltaARS   *int64             `json:"delta_ars"`
	DeltaPct   *float64           `json:"delta_pct"`
	Market     DealMarketSnapshot `json:"market"`
	Disclaimer string             `json:"disclaimer"`
}
