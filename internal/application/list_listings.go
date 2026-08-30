package application

import (
	"context"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type ListListings struct {
	listingRepository domain.ListingRepository
}

func NewListListings(listingRepository domain.ListingRepository) *ListListings {
	return &ListListings{listingRepository: listingRepository}
}

func (u *ListListings) Execute(ctx context.Context, trimID, year, limit int) ([]domain.Listing, error) {
	return u.listingRepository.ListByTrimYear(ctx, trimID, year, limit)
}
