package application

import (
	"context"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type SearchTrims struct {
	vehicleRepository domain.VehicleRepository
}

func NewSearchTrims(vehicleRepository domain.VehicleRepository) *SearchTrims {
	return &SearchTrims{vehicleRepository: vehicleRepository}
}

func (u *SearchTrims) Execute(ctx context.Context, query string, limit int) ([]domain.TrimSearchResult, error) {
	return u.vehicleRepository.SearchTrims(ctx, query, limit)
}
