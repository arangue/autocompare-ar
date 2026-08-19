package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type VehicleRepository struct {
	pool *pgxpool.Pool
}

func NewVehicleRepository(pool *pgxpool.Pool) *VehicleRepository {
	return &VehicleRepository{pool: pool}
}

func (r *VehicleRepository) ListBrands(ctx context.Context) ([]domain.Brand, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, created_at
		FROM brands
		ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	brands := make([]domain.Brand, 0)
	for rows.Next() {
		var b domain.Brand
		if err := rows.Scan(&b.ID, &b.Name, &b.CreatedAt); err != nil {
			return nil, err
		}
		brands = append(brands, b)
	}
	return brands, rows.Err()
}

func (r *VehicleRepository) ListModelsByBrand(_ context.Context, _ int) ([]domain.Model, error) {
	return []domain.Model{}, nil
}

func (r *VehicleRepository) SearchTrims(_ context.Context, _ string, _ int) ([]domain.Trim, error) {
	return []domain.Trim{}, nil
}
