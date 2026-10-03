package mercadolibre

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/arangue/autocompare-ar/internal/domain"
)

// rawItem is one element of search results[]. The field names follow
// docs/ml-api-notes.md (Mercado Libre docs shape, not a live capture).
type rawItem struct {
	ID            string      `json:"id"`
	Title         string      `json:"title"`
	Price         float64     `json:"price"`
	CurrencyID    string      `json:"currency_id"`
	Permalink     string      `json:"permalink"`
	Attributes    []attribute `json:"attributes"`
	Address       *place      `json:"address"`
	SellerAddress *place      `json:"seller_address"`
}

type attribute struct {
	ID        string `json:"id"`
	ValueName string `json:"value_name"`
}

type place struct {
	CityName  string `json:"city_name"`
	StateName string `json:"state_name"`
}

// MapItem converts one Mercado Libre item into a listing.
// ok is false when id, title, year, or price is missing. TrimID stays 0;
// the worker resolves it from raw_title. KM may be nil.
func MapItem(item rawItem) (domain.Listing, bool) {
	id := strings.TrimSpace(item.ID)
	title := strings.TrimSpace(item.Title)
	if id == "" || title == "" || item.Price <= 0 {
		return domain.Listing{}, false
	}
	year, ok := attrInt(item.Attributes, "VEHICLE_YEAR")
	if !ok || year <= 0 {
		return domain.Listing{}, false
	}

	currency := strings.TrimSpace(item.CurrencyID)
	if currency == "" {
		currency = "ARS"
	}

	var km *int
	if n, ok := attrInt(item.Attributes, "KILOMETERS"); ok && n >= 0 {
		km = &n
	}

	return domain.Listing{
		Source:     "mercadolibre",
		ExternalID: id,
		RawTitle:   title,
		Year:       year,
		KM:         km,
		Price:      int64(item.Price),
		Currency:   currency,
		Location:   location(item),
		URL:        strPtr(item.Permalink),
	}, true
}

func attrInt(attrs []attribute, id string) (int, bool) {
	for _, a := range attrs {
		if a.ID != id {
			continue
		}
		digits := strings.Map(func(r rune) rune {
			if unicode.IsDigit(r) {
				return r
			}
			return -1
		}, a.ValueName)
		if digits == "" {
			return 0, false
		}
		n, err := strconv.Atoi(digits)
		if err != nil {
			return 0, false
		}
		return n, true
	}
	return 0, false
}

func location(item rawItem) *string {
	for _, p := range []*place{item.Address, item.SellerAddress} {
		if p == nil {
			continue
		}
		city := strings.TrimSpace(p.CityName)
		state := strings.TrimSpace(p.StateName)
		switch {
		case city != "" && state != "":
			s := city + ", " + state
			return &s
		case city != "":
			return &city
		case state != "":
			return &state
		}
	}
	return nil
}

func strPtr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}
