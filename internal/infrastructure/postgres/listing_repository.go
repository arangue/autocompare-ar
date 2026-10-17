package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arangue/autocompare-ar/internal/domain"
)

var ErrUnknownTrim = errors.New("unknown trim_id")

type ListingRepository struct {
	pool        *pgxpool.Pool
	excludeSeed bool
}

func NewListingRepository(pool *pgxpool.Pool, excludeSeed bool) *ListingRepository {
	return &ListingRepository{pool: pool, excludeSeed: excludeSeed}
}

// activeWhere is the shared filter for the listings list.
// excludeSeed adds source <> 'seed' so deal and listings drop demo rows together.
func (r *ListingRepository) activeWhere() string {
	where := "trim_id = $1 AND year = $2 AND active = true"
	if r.excludeSeed {
		where += " AND source <> 'seed'"
	}
	return where
}

// marketWhere is the used-car median: same rows as the list, minus 0 km.
func (r *ListingRepository) marketWhere() string {
	return r.activeWhere() + " AND (km IS NULL OR km > 0)"
}

func (r *ListingRepository) usedActivePred(alias string) string {
	where := alias + ".active = true AND (" + alias + ".km IS NULL OR " + alias + ".km > 0)"
	if r.excludeSeed {
		where += " AND " + alias + ".source <> 'seed'"
	}
	return where
}

func (r *ListingRepository) ListByTrimYear(ctx context.Context, trimID, year, limit int, currency string) ([]domain.Listing, error) {
	query := `
		SELECT id, source, external_id, trim_id, year, km, price, currency, location, url, last_seen_at
		FROM vehicle_listings
		WHERE ` + r.activeWhere()
	args := []any{trimID, year}
	if currency != "" {
		query += `
		  AND currency = $3
		ORDER BY price ASC
		LIMIT $4`
		args = append(args, currency, limit)
	} else {
		query += `
		ORDER BY price ASC
		LIMIT $3`
		args = append(args, limit)
	}
	rows, err := r.pool.Query(ctx, query, args...)
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

func (r *ListingRepository) currencyPred() string {
	return `currency = COALESCE(NULLIF($3, ''), (
		SELECT currency FROM vehicle_listings
		WHERE ` + r.marketWhere() + `
		GROUP BY currency
		ORDER BY COUNT(*) DESC, CASE WHEN currency = 'ARS' THEN 0 ELSE 1 END, currency
		LIMIT 1
	))`
}

func (r *ListingRepository) MarketSummary(ctx context.Context, trimID, year int, currency string) (domain.MarketSummary, error) {
	var summary domain.MarketSummary
	err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*)::int,
			MIN(price),
			MAX(price),
			(percentile_cont(0.25) WITHIN GROUP (ORDER BY price))::bigint,
			(percentile_cont(0.5) WITHIN GROUP (ORDER BY price))::bigint,
			(percentile_cont(0.75) WITHIN GROUP (ORDER BY price))::bigint,
			COALESCE(MIN(currency), 'ARS'),
			MAX(last_seen_at)
		FROM vehicle_listings
		WHERE `+r.marketWhere()+`
		  AND `+r.currencyPred(), trimID, year, currency).Scan(
		&summary.Count,
		&summary.Minimum,
		&summary.Maximum,
		&summary.P25,
		&summary.Median,
		&summary.P75,
		&summary.Currency,
		&summary.ObservedAt,
	)
	if err != nil {
		return domain.MarketSummary{}, err
	}
	if summary.Currency == "" {
		summary.Currency = "ARS"
	}
	summary.Sources = []domain.MarketSource{}
	if summary.Count == 0 {
		return summary, nil
	}

	rows, err := r.pool.Query(ctx, `
		SELECT source, COUNT(*)::int
		FROM vehicle_listings
		WHERE `+r.marketWhere()+`
		  AND `+r.currencyPred()+`
		GROUP BY source
		ORDER BY COUNT(*) DESC, source`, trimID, year, currency)
	if err != nil {
		return domain.MarketSummary{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var src domain.MarketSource
		if err := rows.Scan(&src.Source, &src.Count); err != nil {
			return domain.MarketSummary{}, err
		}
		summary.Sources = append(summary.Sources, src)
	}
	if err := rows.Err(); err != nil {
		return domain.MarketSummary{}, err
	}
	return summary, nil
}

func (r *ListingRepository) ListModelOffer(ctx context.Context) (domain.ModelOfferList, error) {
	list := domain.ModelOfferList{Models: []domain.ModelOffer{}}
	err := r.pool.QueryRow(ctx, `
		SELECT MAX(last_seen_at)
		FROM vehicle_listings vl
		WHERE `+r.usedActivePred("vl")).Scan(&list.ObservedAt)
	if err != nil {
		return domain.ModelOfferList{}, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT m.id, m.name, b.id, b.name, COUNT(*)::int
		FROM vehicle_listings vl
		JOIN trims t ON t.id = vl.trim_id
		JOIN generations g ON g.id = t.generation_id
		JOIN models m ON m.id = g.model_id
		JOIN brands b ON b.id = m.brand_id
		WHERE `+r.usedActivePred("vl")+`
		GROUP BY m.id, m.name, b.id, b.name
		ORDER BY COUNT(*) DESC, b.name, m.name`)
	if err != nil {
		return domain.ModelOfferList{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var model domain.ModelOffer
		if err := rows.Scan(&model.ModelID, &model.ModelName, &model.BrandID, &model.BrandName, &model.Count); err != nil {
			return domain.ModelOfferList{}, err
		}
		list.Models = append(list.Models, model)
	}
	if err := rows.Err(); err != nil {
		return domain.ModelOfferList{}, err
	}
	return list, nil
}

func (r *ListingRepository) GetModelOffer(ctx context.Context, modelID int) (domain.ModelOfferDetail, error) {
	var detail domain.ModelOfferDetail
	err := r.pool.QueryRow(ctx, `
		SELECT m.id, m.name, b.id, b.name
		FROM models m
		JOIN brands b ON b.id = m.brand_id
		WHERE m.id = $1`, modelID).Scan(&detail.ModelID, &detail.ModelName, &detail.BrandID, &detail.BrandName)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ModelOfferDetail{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.ModelOfferDetail{}, err
	}
	detail.Trims = []domain.ModelOfferTrim{}

	rows, err := r.pool.Query(ctx, `
		SELECT t.id, t.name, g.name, vl.year, COUNT(*)::int, MAX(vl.last_seen_at)
		FROM vehicle_listings vl
		JOIN trims t ON t.id = vl.trim_id
		JOIN generations g ON g.id = t.generation_id
		WHERE g.model_id = $1
		  AND `+r.usedActivePred("vl")+`
		GROUP BY t.id, t.name, g.name, vl.year
		ORDER BY t.name, vl.year DESC`, modelID)
	if err != nil {
		return domain.ModelOfferDetail{}, err
	}
	defer rows.Close()

	order := make([]int, 0)
	byID := make(map[int]*domain.ModelOfferTrim)
	for rows.Next() {
		var trimID int
		var trimName, generation string
		var year, count int
		var seen time.Time
		if err := rows.Scan(&trimID, &trimName, &generation, &year, &count, &seen); err != nil {
			return domain.ModelOfferDetail{}, err
		}
		detail.Count += count
		if detail.ObservedAt == nil || seen.After(*detail.ObservedAt) {
			seen := seen
			detail.ObservedAt = &seen
		}
		trim, ok := byID[trimID]
		if !ok {
			trim = &domain.ModelOfferTrim{
				TrimID:         trimID,
				TrimName:       trimName,
				GenerationName: generation,
				Years:          []domain.ModelOfferYear{},
			}
			byID[trimID] = trim
			order = append(order, trimID)
		}
		trim.Count += count
		trim.Years = append(trim.Years, domain.ModelOfferYear{Year: year, Count: count})
	}
	if err := rows.Err(); err != nil {
		return domain.ModelOfferDetail{}, err
	}
	for _, id := range order {
		detail.Trims = append(detail.Trims, *byID[id])
	}
	return detail, nil
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
	return r.expire(ctx, "last_seen_at < NOW() - make_interval(days => $1)", []any{days}, "")
}

// ExpireUnseen deactivates listings last seen before cutoff.
// Empty source updates every source.
func (r *ListingRepository) ExpireUnseen(ctx context.Context, before time.Time, source string) (int64, error) {
	if before.IsZero() {
		return 0, nil
	}
	return r.expire(ctx, "last_seen_at < $1", []any{before}, source)
}

func (r *ListingRepository) expire(ctx context.Context, pred string, args []any, source string) (int64, error) {
	query := `UPDATE vehicle_listings SET active = false WHERE active AND ` + pred
	if source != "" {
		query += fmt.Sprintf(` AND source = $%d`, len(args)+1)
		args = append(args, source)
	}
	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
