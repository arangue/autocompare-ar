package application

import (
	"context"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type ListBrands interface {
	Execute(ctx context.Context) ([]domain.Brand, error)
}

type listBrands struct {
	repo domain.VehicleRepository
}

func NewListBrands(repo domain.VehicleRepository) ListBrands {
	return &listBrands{repo: repo}
}

func (uc *listBrands) Execute(ctx context.Context) ([]domain.Brand, error) {
	return uc.repo.ListBrands(ctx)
}
