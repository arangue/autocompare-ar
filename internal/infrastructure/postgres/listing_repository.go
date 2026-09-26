package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arangue/autocompare-ar/internal/domain"
)

var ErrUnknownTrim = errors.New("unknown trim_id")

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

func (r *ListingRepository) Upsert(ctx context.Context, listing domain.Listing) (bool, error) {
	var lastSeen *time.Time
	if !listing.LastSeenAt.IsZero() {
		t := listing.LastSeenAt
		lastSeen = &t
	}
	var inserted bool
	err := r.pool.QueryRow(ctx, `
		INSERT INTO vehicle_listings (
			source, external_id, trim_id, year, km, price, currency, location, url, last_seen_at, active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, COALESCE($10, NOW()), COALESCE($11, TRUE))
		ON CONFLICT (source, external_id) DO UPDATE SET
			trim_id = EXCLUDED.trim_id,
			year = EXCLUDED.year,
			km = EXCLUDED.km,
			price = EXCLUDED.price,
			currency = EXCLUDED.currency,
			location = EXCLUDED.location,
			url = EXCLUDED.url,
			last_seen_at = COALESCE($10, NOW()),
			active = COALESCE($11, TRUE)
		RETURNING (xmax = 0)`,
		listing.Source, listing.ExternalID, listing.TrimID, listing.Year,
		listing.KM, listing.Price, listing.Currency, listing.Location, listing.URL,
		lastSeen, listing.Active,
	).Scan(&inserted)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return false, ErrUnknownTrim
		}
		return false, err
	}
	return inserted, nil
}

func (r *ListingRepository) ExpireStale(ctx context.Context, days int) (int64, error) {
	if days <= 0 {
		return 0, nil
	}
	tag, err := r.pool.Exec(ctx, `
		UPDATE vehicle_listings
		SET active = false
		WHERE active AND last_seen_at < NOW() - make_interval(days => $1)`, days)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
