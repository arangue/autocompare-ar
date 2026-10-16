package mercadolibre

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/arangue/autocompare-ar/internal/domain"
)

var mlaID = regexp.MustCompile(`(?i)MLA-?(\d+)`)

// ParseSearchFile reads a scraper CSV into listings. TrimID stays 0.
// skipped counts rows with no MLA id, year, or price.
func ParseSearchFile(path string) ([]domain.Listing, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, fmt.Errorf("mercadolibre csv: %w", err)
	}
	return ParseSearchCSV(data)
}

// ParseSearchCSV reads Title, Price, Currency, Year, Kilometers, Location, URL.
// Older files used "Price (ARS)" and no Currency column; those default to ARS.
func ParseSearchCSV(data []byte) ([]domain.Listing, int, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, 0, fmt.Errorf("mercadolibre csv: %w", err)
	}
	if len(records) == 0 {
		return nil, 0, nil
	}

	idx := map[string]int{}
	for i, header := range records[0] {
		idx[strings.ToLower(strings.TrimSpace(header))] = i
	}
	if _, ok := priceColumn(idx); !ok {
		return nil, 0, fmt.Errorf("mercadolibre csv missing column %q", "price")
	}
	for _, required := range []string{"title", "year", "url"} {
		if _, ok := idx[required]; !ok {
			return nil, 0, fmt.Errorf("mercadolibre csv missing column %q", required)
		}
	}

	listings := make([]domain.Listing, 0, len(records)-1)
	skipped := 0
	for _, record := range records[1:] {
		listing, ok := searchRow(record, idx)
		if !ok {
			skipped++
			continue
		}
		listings = append(listings, listing)
	}
	return listings, skipped, nil
}

func searchRow(record []string, idx map[string]int) (domain.Listing, bool) {
	title := csvField(record, idx, "title")
	rawURL := csvField(record, idx, "url")
	id, ok := ExternalID(rawURL)
	if !ok || title == "" {
		return domain.Listing{}, false
	}
	year, err := strconv.Atoi(digitsOnly(csvField(record, idx, "year")))
	if err != nil || year <= 0 {
		return domain.Listing{}, false
	}
	priceKey, ok := priceColumn(idx)
	if !ok {
		return domain.Listing{}, false
	}
	price, err := strconv.ParseInt(digitsOnly(csvField(record, idx, priceKey)), 10, 64)
	if err != nil || price <= 0 {
		return domain.Listing{}, false
	}

	listing := domain.Listing{
		Source:     "mercadolibre",
		ExternalID: id,
		RawTitle:   title,
		Year:       year,
		Price:      price,
		Currency:   listingCurrency(csvField(record, idx, "currency")),
		URL:        strPtr(rawURL),
		Location:   strPtr(csvField(record, idx, "location")),
	}
	if kmText := digitsOnly(csvField(record, idx, "kilometers")); kmText != "" {
		km, err := strconv.Atoi(kmText)
		if err != nil {
			return domain.Listing{}, false
		}
		listing.KM = &km
	}
	return listing, true
}

// ExternalID reads MLA plus the item digits. MLA-2168520435 becomes MLA2168520435.
func ExternalID(rawURL string) (string, bool) {
	m := mlaID.FindStringSubmatch(rawURL)
	if m == nil {
		return "", false
	}
	return "MLA" + m[1], true
}

func priceColumn(idx map[string]int) (string, bool) {
	if _, ok := idx["price"]; ok {
		return "price", true
	}
	if _, ok := idx["price (ars)"]; ok {
		return "price (ars)", true
	}
	return "", false
}

func listingCurrency(raw string) string {
	s := strings.ToUpper(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, " ", "")
	switch s {
	case "USD", "US$", "U$S", "U$D":
		return "USD"
	default:
		return "ARS"
	}
}

func digitsOnly(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, s)
}

func csvField(record []string, idx map[string]int, key string) string {
	i, ok := idx[key]
	if !ok || i >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[i])
}
