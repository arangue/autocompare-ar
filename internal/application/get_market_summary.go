package application

import (
	"context"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type GetMarketSummary struct {
	listingRepository domain.ListingRepository
}

func NewGetMarketSummary(listingRepository domain.ListingRepository) *GetMarketSummary {
	return &GetMarketSummary{listingRepository: listingRepository}
}

func (u *GetMarketSummary) Execute(ctx context.Context, trimID, year int) (domain.TrimMarketSummary, error) {
	summary, err := u.listingRepository.MarketSummary(ctx, trimID, year)
	if err != nil {
		return domain.TrimMarketSummary{}, err
	}
	return domain.TrimMarketSummary{
		TrimID:   trimID,
		Year:     year,
		Count:    summary.Count,
		Median:   summary.Median,
		Minimum:  summary.Minimum,
		Maximum:  summary.Maximum,
		P25:      summary.P25,
		P75:      summary.P75,
		Currency: "ARS",
	}, nil
}
