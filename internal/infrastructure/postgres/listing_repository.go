package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type ListingRepository struct {
	pool *pgxpool.Pool
}

func NewListingRepository(pool *pgxpool.Pool) *ListingRepository {
	return &ListingRepository{pool: pool}
}

func (r *ListingRepository) ListByTrimYear(ctx context.Context, trimID, year, limit int) ([]domain.Listing, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, source, external_id, trim_id, year, km, price, currency, location, url, last_seen_at
		FROM vehicle_listings
		WHERE trim_id = $1 AND year = $2 AND active = true
		ORDER BY price ASC
		LIMIT $3`, trimID, year, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	listings := make([]domain.Listing, 0)
	for rows.Next() {
		var l domain.Listing
		if err := rows.Scan(
			&l.ID, &l.Source, &l.ExternalID, &l.TrimID, &l.Year,
			&l.KM, &l.Price, &l.Currency, &l.Location, &l.URL, &l.LastSeenAt,
		); err != nil {
			return nil, err
		}
		listings = append(listings, l)
	}
	return listings, rows.Err()
}

func (r *ListingRepository) MarketSummary(ctx context.Context, trimID, year int) (domain.MarketSummary, error) {
	var summary domain.MarketSummary
	err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*)::int,
			MIN(price),
			MAX(price),
			(percentile_cont(0.25) WITHIN GROUP (ORDER BY price))::bigint,
			(percentile_cont(0.5) WITHIN GROUP (ORDER BY price))::bigint,
			(percentile_cont(0.75) WITHIN GROUP (ORDER BY price))::bigint
		FROM vehicle_listings
		WHERE trim_id = $1 AND year = $2 AND active = true`, trimID, year).Scan(
		&summary.Count,
		&summary.Minimum,
		&summary.Maximum,
		&summary.P25,
		&summary.Median,
		&summary.P75,
	)
	if err != nil {
		return domain.MarketSummary{}, err
	}
	return summary, nil
}
