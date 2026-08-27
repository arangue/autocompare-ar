package application

import (
	"context"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type ListModelsByBrand interface {
	Execute(ctx context.Context, brandID int) ([]domain.Model, error)
}

type listModelsByBrand struct {
	repo domain.VehicleRepository
}

func NewListModelsByBrand(repo domain.VehicleRepository) ListModelsByBrand {
	return &listModelsByBrand{repo: repo}
}

func (uc *listModelsByBrand) Execute(ctx context.Context, brandID int) ([]domain.Model, error) {
	return uc.repo.ListModelsByBrand(ctx, brandID)
}
