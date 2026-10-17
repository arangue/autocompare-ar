package application

import (
	"context"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type GetModelOffer struct {
	listingRepository domain.ListingRepository
}

func NewGetModelOffer(listingRepository domain.ListingRepository) *GetModelOffer {
	return &GetModelOffer{listingRepository: listingRepository}
}

func (u *GetModelOffer) Execute(ctx context.Context, modelID int) (domain.ModelOfferDetail, error) {
	return u.listingRepository.GetModelOffer(ctx, modelID)
}
