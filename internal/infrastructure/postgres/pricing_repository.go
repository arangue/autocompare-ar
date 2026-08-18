package postgres

import (
	"context"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type PricingRepository struct{}

func NewPricingRepository() *PricingRepository {
	return &PricingRepository{}
}

func (r *PricingRepository) ListReferences(_ context.Context, _, _ int) ([]domain.ReferencePrice, error) {
	return []domain.ReferencePrice{}, nil
}
