package ingestion

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/arangue/autocompare-ar/internal/domain"
)

// ParseFile reads JSON (array) or CSV listings. Invalid rows are skipped.
func ParseFile(path string) (listings []domain.Listing, skipped int, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return parseJSON(data)
	case ".csv":
		return parseCSV(data)
	default:
		return nil, 0, fmt.Errorf("unsupported ingest file type %q", filepath.Ext(path))
	}
}

type fileRow struct {
	Source     string  `json:"source"`
	ExternalID string  `json:"external_id"`
	TrimID     int     `json:"trim_id"`
	RawTitle   string  `json:"raw_title"`
	Year       int     `json:"year"`
	KM         *int    `json:"km"`
	Price      int64   `json:"price"`
	Currency   string  `json:"currency"`
	Location   *string `json:"location"`
	URL        *string `json:"url"`
}

func parseJSON(data []byte) ([]domain.Listing, int, error) {
	var rows []fileRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, 0, fmt.Errorf("parse json listings: %w", err)
	}

	listings := make([]domain.Listing, 0, len(rows))
	skipped := 0
	for _, row := range rows {
		listing, ok := row.toListing()
		if !ok {
			skipped++
			continue
		}
		listings = append(listings, listing)
	}
	return listings, skipped, nil
}

func parseCSV(data []byte) ([]domain.Listing, int, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, 0, fmt.Errorf("parse csv listings: %w", err)
	}
	if len(records) == 0 {
		return nil, 0, nil
	}

	idx := map[string]int{}
	for i, header := range records[0] {
		idx[strings.ToLower(strings.TrimSpace(header))] = i
	}
	for _, required := range []string{"source", "external_id", "trim_id", "year", "price"} {
		if _, ok := idx[required]; !ok {
			return nil, 0, fmt.Errorf("csv missing column %q", required)
		}
	}

	listings := make([]domain.Listing, 0, len(records)-1)
	skipped := 0
	for _, record := range records[1:] {
		row := fileRow{
			Source:     csvField(record, idx, "source"),
			ExternalID: csvField(record, idx, "external_id"),
			Currency:   csvField(record, idx, "currency"),
		}
		row.TrimID, _ = strconv.Atoi(csvField(record, idx, "trim_id"))
		row.RawTitle = csvField(record, idx, "raw_title")
		row.Year, _ = strconv.Atoi(csvField(record, idx, "year"))
		row.Price, _ = strconv.ParseInt(csvField(record, idx, "price"), 10, 64)
		if km := csvField(record, idx, "km"); km != "" {
			if n, err := strconv.Atoi(km); err == nil {
				row.KM = &n
			}
		}
		if loc := csvField(record, idx, "location"); loc != "" {
			row.Location = &loc
		}
		if url := csvField(record, idx, "url"); url != "" {
			row.URL = &url
		}

		listing, ok := row.toListing()
		if !ok {
			skipped++
			continue
		}
		listings = append(listings, listing)
	}
	return listings, skipped, nil
}

func csvField(record []string, idx map[string]int, key string) string {
	i, ok := idx[key]
	if !ok || i >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[i])
}

func (r fileRow) toListing() (domain.Listing, bool) {
	if strings.TrimSpace(r.Source) == "" || strings.TrimSpace(r.ExternalID) == "" {
		return domain.Listing{}, false
	}
	rawTitle := strings.TrimSpace(r.RawTitle)
	if r.Year <= 0 || r.Price <= 0 {
		return domain.Listing{}, false
	}
	if r.TrimID <= 0 && rawTitle == "" {
		return domain.Listing{}, false
	}

	currency := strings.TrimSpace(r.Currency)
	if currency == "" {
		currency = "ARS"
	}

	return domain.Listing{
		Source:     strings.TrimSpace(r.Source),
		ExternalID: strings.TrimSpace(r.ExternalID),
		TrimID:     r.TrimID,
		RawTitle:   rawTitle,
		Year:       r.Year,
		KM:         r.KM,
		Price:      r.Price,
		Currency:   currency,
		Location:   emptyNil(r.Location),
		URL:        emptyNil(r.URL),
	}, true
}

func emptyNil(s *string) *string {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	return s
}
