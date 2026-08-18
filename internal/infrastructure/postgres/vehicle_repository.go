package postgres

import (
	"context"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type VehicleRepository struct{}

func NewVehicleRepository() *VehicleRepository {
	return &VehicleRepository{}
}

func (r *VehicleRepository) ListBrands(_ context.Context) ([]domain.Brand, error) {
	return []domain.Brand{}, nil
}

func (r *VehicleRepository) ListModelsByBrand(_ context.Context, _ int) ([]domain.Model, error) {
	return []domain.Model{}, nil
}

func (r *VehicleRepository) SearchTrims(_ context.Context, _ string, _ int) ([]domain.Trim, error) {
	return []domain.Trim{}, nil
}
