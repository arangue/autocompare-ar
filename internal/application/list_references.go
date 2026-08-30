package application

import (
	"context"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type ListReferences struct {
	repository domain.PricingRepository
}

func NewListReferences(repository domain.PricingRepository) *ListReferences {
	return &ListReferences{repository: repository}
}

func (u *ListReferences) Execute(ctx context.Context, trimID, year int) ([]domain.ReferencePrice, error) {
	return u.repository.ListReferences(ctx, trimID, year)
}
