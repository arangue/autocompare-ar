package application

import (
	"context"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type GetTrim struct {
	vehicleRepository domain.VehicleRepository
}

func NewGetTrimService(vehicleRepository domain.VehicleRepository) *GetTrim {
	return &GetTrim{vehicleRepository: vehicleRepository}
}

func (u *GetTrim) Execute(ctx context.Context, id int) (domain.TrimDetail, error) {
	return u.vehicleRepository.GetTrim(ctx, id)
}
