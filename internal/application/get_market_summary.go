package application

import (
	"context"
	"strings"

	"github.com/arangue/autocompare-ar/internal/domain"
)

func listingCurrency(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return "ARS"
	}
	return raw
}

type GetMarketSummary struct {
	listingRepository domain.ListingRepository
}

func NewGetMarketSummary(listingRepository domain.ListingRepository) *GetMarketSummary {
	return &GetMarketSummary{listingRepository: listingRepository}
}

func (u *GetMarketSummary) Execute(ctx context.Context, trimID, year int, currency string) (domain.TrimMarketSummary, error) {
	summary, err := u.listingRepository.MarketSummary(ctx, trimID, year, currency)
	if err != nil {
		return domain.TrimMarketSummary{}, err
	}
	sources := summary.Sources
	if sources == nil {
		sources = []domain.MarketSource{}
	}
	return domain.TrimMarketSummary{
		TrimID:     trimID,
		Year:       year,
		Count:      summary.Count,
		Median:     summary.Median,
		Minimum:    summary.Minimum,
		Maximum:    summary.Maximum,
		P25:        summary.P25,
		P75:        summary.P75,
		Currency:   listingCurrency(summary.Currency),
		ObservedAt: summary.ObservedAt,
		Sources:    sources,
	}, nil
}
