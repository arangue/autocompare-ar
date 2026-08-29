package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
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

func (r *VehicleRepository) ListModelsByBrand(ctx context.Context, brandID int) ([]domain.Model, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, brand_id, name, created_at
		FROM models
		WHERE brand_id = $1
		ORDER BY name`, brandID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	models := make([]domain.Model, 0)
	for rows.Next() {
		var m domain.Model
		if err := rows.Scan(&m.ID, &m.BrandID, &m.Name, &m.CreatedAt); err != nil {
			return nil, err
		}
		models = append(models, m)
	}
	return models, rows.Err()
}

func (r *VehicleRepository) SearchTrims(ctx context.Context, query string, limit int) ([]domain.TrimSearchResult, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			t.id,
			t.name,
			b.name,
			m.name,
			g.name,
			t.year_from,
			t.year_to
		FROM trims t
		JOIN generations g ON g.id = t.generation_id
		JOIN models m      ON m.id = g.model_id
		JOIN brands b      ON b.id = m.brand_id
		WHERE b.name ILIKE '%' || $1 || '%'
		   OR m.name ILIKE '%' || $1 || '%'
		   OR t.name ILIKE '%' || $1 || '%'
		ORDER BY b.name, m.name, t.name
		LIMIT $2`, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trims := make([]domain.TrimSearchResult, 0)
	for rows.Next() {
		var t domain.TrimSearchResult
		if err := rows.Scan(&t.TrimID, &t.TrimName, &t.BrandName, &t.ModelName, &t.GenerationName, &t.YearFrom, &t.YearTo); err != nil {
			return nil, err
		}
		trims = append(trims, t)
	}
	return trims, rows.Err()
}

func (r *VehicleRepository) GetTrim(ctx context.Context, id int) (domain.TrimDetail, error) {
	var d domain.TrimDetail
	query := `SELECT
				t.id,
				t.name,
				b.id,
				b.name,
				m.id,
				m.name,
				g.id,
				g.name,
				t.year_from,
				t.year_to
			FROM trims t
			JOIN generations g ON g.id = t.generation_id
			JOIN models m      ON m.id = g.model_id
			JOIN brands b      ON b.id = m.brand_id
			WHERE t.id = $1`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&d.TrimID, &d.TrimName,
		&d.BrandID, &d.BrandName,
		&d.ModelID, &d.ModelName,
		&d.GenerationID, &d.GenerationName,
		&d.YearFrom, &d.YearTo,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TrimDetail{}, domain.ErrNotFound
		}
		return domain.TrimDetail{}, err
	}

	d.Features = []domain.VehicleFeature{}
	if err := r.loadTrimSpecs(ctx, id, &d); err != nil {
		return domain.TrimDetail{}, err
	}
	if err := r.loadTrimFeatures(ctx, id, &d); err != nil {
		return domain.TrimDetail{}, err
	}
	return d, nil
}

func (r *VehicleRepository) loadTrimSpecs(ctx context.Context, trimID int, d *domain.TrimDetail) error {
	var s domain.VehicleSpec
	err := r.pool.QueryRow(ctx, `
		SELECT engine, displacement_cc, horsepower, transmission, fuel,
		       doors, seats, consumption_city, consumption_highway
		FROM vehicle_specs
		WHERE trim_id = $1`, trimID).Scan(
		&s.Engine, &s.DisplacementCC, &s.Horsepower, &s.Transmission, &s.Fuel,
		&s.Doors, &s.Seats, &s.ConsumptionCity, &s.ConsumptionHighway,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	d.Specs = &s
	return nil
}

func (r *VehicleRepository) loadTrimFeatures(ctx context.Context, trimID int, d *domain.TrimDetail) error {
	rows, err := r.pool.Query(ctx, `
		SELECT f.code, f.name, f.category, vf.value
		FROM vehicle_features vf
		JOIN features f ON f.id = vf.feature_id
		WHERE vf.trim_id = $1
		ORDER BY f.category, f.name`, trimID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var f domain.VehicleFeature
		if err := rows.Scan(&f.Code, &f.Name, &f.Category, &f.Value); err != nil {
			return err
		}
		d.Features = append(d.Features, f)
	}
	return rows.Err()
}
