package application

import (
	"context"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type ListModelOffer struct {
	listingRepository domain.ListingRepository
}

func NewListModelOffer(listingRepository domain.ListingRepository) *ListModelOffer {
	return &ListModelOffer{listingRepository: listingRepository}
}

func (u *ListModelOffer) Execute(ctx context.Context) (domain.ModelOfferList, error) {
	list, err := u.listingRepository.ListModelOffer(ctx)
	if err != nil {
		return domain.ModelOfferList{}, err
	}
	if list.Models == nil {
		list.Models = []domain.ModelOffer{}
	}
	return list, nil
}
