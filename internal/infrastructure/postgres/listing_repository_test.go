package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arangue/autocompare-ar/internal/domain"
)

func fixtureTrim(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()
	ctx := context.Background()
	var id int
	err := pool.QueryRow(ctx, `
		WITH b AS (
			INSERT INTO brands (name) VALUES ('zz-test')
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		), m AS (
			INSERT INTO models (brand_id, name)
			SELECT id, 'zz-test' FROM b
			ON CONFLICT (brand_id, name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		), g AS (
			INSERT INTO generations (model_id, name, year_from)
			SELECT id, 'zz-test', 2019 FROM m
			ON CONFLICT (model_id, name) DO UPDATE SET year_from = EXCLUDED.year_from
			RETURNING id
		)
		INSERT INTO trims (generation_id, name, year_from)
		SELECT id, 'zz-test', 2019 FROM g
		ON CONFLICT (generation_id, name) DO UPDATE SET year_from = EXCLUDED.year_from
		RETURNING id`).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM vehicle_listings WHERE trim_id = $1`, id)
		_, _ = pool.Exec(context.Background(), `DELETE FROM trims WHERE id = $1`, id)
		_, _ = pool.Exec(context.Background(), `DELETE FROM generations WHERE name = 'zz-test'`)
		_, _ = pool.Exec(context.Background(), `DELETE FROM models WHERE name = 'zz-test'`)
		_, _ = pool.Exec(context.Background(), `DELETE FROM brands WHERE name = 'zz-test'`)
	})
	return id
}

func TestUpsert_activeFalsePersists(t *testing.T) {
	pool := catalogPool(t)
	defer pool.Close()
	repo := NewListingRepository(pool, false)
	ctx := context.Background()
	trimID := fixtureTrim(t, pool)

	off := false
	listing := domain.Listing{
		Source:     "file",
		ExternalID: "e4-402-inactive",
		TrimID:     trimID,
		Year:       2019,
		Price:      1,
		Currency:   "ARS",
		Active:     &off,
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM vehicle_listings WHERE source = $1 AND external_id = $2`,
			listing.Source, listing.ExternalID)
	})

	if _, err := repo.Upsert(ctx, listing); err != nil {
		t.Fatal(err)
	}

	var active bool
	err := pool.QueryRow(ctx,
		`SELECT active FROM vehicle_listings WHERE source = $1 AND external_id = $2`,
		listing.Source, listing.ExternalID,
	).Scan(&active)
	if err != nil {
		t.Fatal(err)
	}
	if active {
		t.Fatal("active:false was not persisted")
	}

	got, err := repo.ListByTrimYear(ctx, listing.TrimID, listing.Year, 100, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range got {
		if l.ExternalID == listing.ExternalID {
			t.Fatal("inactive listing still in market list")
		}
	}
}

func TestExpireStale_noop(t *testing.T) {
	n, err := (&ListingRepository{}).ExpireStale(context.Background(), 0)
	if err != nil || n != 0 {
		t.Fatalf("n=%d err=%v", n, err)
	}
}

func TestExpireStale_deactivatesOld(t *testing.T) {
	pool := catalogPool(t)
	defer pool.Close()
	repo := NewListingRepository(pool, false)
	ctx := context.Background()
	trimID := fixtureTrim(t, pool)

	listing := domain.Listing{
		Source:     "file",
		ExternalID: "e4-403-stale",
		TrimID:     trimID,
		Year:       2019,
		Price:      1,
		Currency:   "ARS",
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM vehicle_listings WHERE source = $1 AND external_id = $2`,
			listing.Source, listing.ExternalID)
	})

	if _, err := repo.Upsert(ctx, listing); err != nil {
		t.Fatal(err)
	}
	_, err := pool.Exec(ctx, `
		UPDATE vehicle_listings
		SET last_seen_at = NOW() - INTERVAL '30 days'
		WHERE source = $1 AND external_id = $2`, listing.Source, listing.ExternalID)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := repo.ExpireStale(ctx, 14); err != nil {
		t.Fatal(err)
	}

	var active bool
	err = pool.QueryRow(ctx,
		`SELECT active FROM vehicle_listings WHERE source = $1 AND external_id = $2`,
		listing.Source, listing.ExternalID,
	).Scan(&active)
	if err != nil {
		t.Fatal(err)
	}
	if active {
		t.Fatal("stale listing still active")
	}

	got, err := repo.ListByTrimYear(ctx, listing.TrimID, listing.Year, 100, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range got {
		if l.ExternalID == listing.ExternalID {
			t.Fatal("stale listing still in market list")
		}
	}
}

func TestActiveWhere(t *testing.T) {
	off := (&ListingRepository{}).activeWhere()
	on := (&ListingRepository{excludeSeed: true}).activeWhere()
	if off != "trim_id = $1 AND year = $2 AND active = true" {
		t.Fatalf("off = %q", off)
	}
	if on != off+" AND source <> 'seed'" {
		t.Fatalf("on = %q", on)
	}
}

func TestExcludeSeed_sameRuleForListAndMarket(t *testing.T) {
	pool := catalogPool(t)
	defer pool.Close()
	ctx := context.Background()
	off := NewListingRepository(pool, false)
	on := NewListingRepository(pool, true)
	trimID := fixtureTrim(t, pool)

	seed := domain.Listing{Source: "seed", ExternalID: "e5-502-seed", TrimID: trimID, Year: 2099, Price: 100, Currency: "ARS"}
	live := domain.Listing{Source: "file", ExternalID: "e5-502-file", TrimID: trimID, Year: 2099, Price: 200, Currency: "ARS"}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM vehicle_listings WHERE external_id IN ('e5-502-seed', 'e5-502-file')`)
	})
	for _, listing := range []domain.Listing{seed, live} {
		if _, err := off.Upsert(ctx, listing); err != nil {
			t.Fatal(err)
		}
	}

	offMarket, err := off.MarketSummary(ctx, trimID, 2099, "")
	if err != nil {
		t.Fatal(err)
	}
	offList, err := off.ListByTrimYear(ctx, trimID, 2099, 100, "")
	if err != nil {
		t.Fatal(err)
	}
	if offMarket.Count != 2 || len(offList) != 2 {
		t.Fatalf("flag off market=%d list=%d", offMarket.Count, len(offList))
	}

	onMarket, err := on.MarketSummary(ctx, trimID, 2099, "")
	if err != nil {
		t.Fatal(err)
	}
	onList, err := on.ListByTrimYear(ctx, trimID, 2099, 100, "")
	if err != nil {
		t.Fatal(err)
	}
	if onMarket.Count != 1 || len(onList) != 1 || onList[0].Source != "file" {
		t.Fatalf("flag on market=%d list=%v", onMarket.Count, onList)
	}
}

func TestMarketSummary_majorityCurrency(t *testing.T) {
	pool := catalogPool(t)
	defer pool.Close()
	repo := NewListingRepository(pool, false)
	ctx := context.Background()
	trimID := fixtureTrim(t, pool)

	rows := []domain.Listing{
		{Source: "file", ExternalID: "cur-usd-1", TrimID: trimID, Year: 2088, Price: 20_000, Currency: "USD"},
		{Source: "file", ExternalID: "cur-usd-2", TrimID: trimID, Year: 2088, Price: 22_000, Currency: "USD"},
		{Source: "file", ExternalID: "cur-ars-1", TrimID: trimID, Year: 2088, Price: 40_000_000, Currency: "ARS"},
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM vehicle_listings WHERE external_id LIKE 'cur-%'`)
	})
	for _, listing := range rows {
		if _, err := repo.Upsert(ctx, listing); err != nil {
			t.Fatal(err)
		}
	}

	summary, err := repo.MarketSummary(ctx, trimID, 2088, "")
	if err != nil {
		t.Fatal(err)
	}
	if summary.Count != 2 || summary.Currency != "USD" {
		t.Fatalf("summary = %+v", summary)
	}
	if summary.Median == nil || *summary.Median != 21_000 {
		t.Fatalf("median = %v", summary.Median)
	}

	all, err := repo.ListByTrimYear(ctx, trimID, 2088, 10, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("all list = %d", len(all))
	}

	list, err := repo.ListByTrimYear(ctx, trimID, 2088, 10, "USD")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("usd list = %d", len(list))
	}
	for _, listing := range list {
		if listing.Currency != "USD" {
			t.Fatalf("mixed list currency = %s", listing.Currency)
		}
	}

	ars, err := repo.MarketSummary(ctx, trimID, 2088, "ARS")
	if err != nil {
		t.Fatal(err)
	}
	if ars.Count != 1 || ars.Currency != "ARS" {
		t.Fatalf("ars summary = %+v", ars)
	}
	if ars.Median == nil || *ars.Median != 40_000_000 {
		t.Fatalf("ars median = %v", ars.Median)
	}
}
