package postgres

import (
	"context"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type ListingRepository struct{}

func NewListingRepository() *ListingRepository {
	return &ListingRepository{}
}

func (r *ListingRepository) ListByTrimYear(_ context.Context, _, _, _ int) ([]domain.Listing, error) {
	return []domain.Listing{}, nil
}

func (r *ListingRepository) MarketSummary(_ context.Context, _, _ int) (domain.MarketSummary, error) {
	return domain.MarketSummary{}, nil
}
