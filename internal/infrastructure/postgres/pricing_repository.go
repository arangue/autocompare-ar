package postgres

import (
	"context"

	"github.com/arangue/autocompare-ar/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PricingRepository struct {
	pool *pgxpool.Pool
}

func NewPricingRepository(pool *pgxpool.Pool) *PricingRepository {
	return &PricingRepository{pool: pool}
}

func (r *PricingRepository) ListReferences(ctx context.Context, trimID, year int) ([]domain.ReferencePrice, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, trim_id, year, source, kind, price, currency, observed_at
		FROM price_references
		WHERE trim_id = $1 AND year = $2
		ORDER BY price ASC`, trimID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	references := make([]domain.ReferencePrice, 0)
	for rows.Next() {
		var reference domain.ReferencePrice
		if err := rows.Scan(&reference.ID, &reference.TrimID, &reference.Year, &reference.Source, &reference.Kind, &reference.Price, &reference.Currency, &reference.ObservedAt); err != nil {
			return nil, err
		}
		references = append(references, reference)
	}
	return references, rows.Err()
}
