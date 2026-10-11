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

// UpsertReference inserts or updates one guide price on (trim_id, year, source).
func (r *PricingRepository) UpsertReference(ctx context.Context, ref domain.ReferencePrice) (inserted bool, err error) {
	err = r.pool.QueryRow(ctx, `
		INSERT INTO price_references (trim_id, year, source, kind, price, currency, observed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (trim_id, year, source) DO UPDATE SET
			kind = EXCLUDED.kind,
			price = EXCLUDED.price,
			currency = EXCLUDED.currency,
			observed_at = EXCLUDED.observed_at
		RETURNING (xmax = 0)`,
		ref.TrimID, ref.Year, ref.Source, ref.Kind, ref.Price, ref.Currency, ref.ObservedAt,
	).Scan(&inserted)
	return inserted, err
}
