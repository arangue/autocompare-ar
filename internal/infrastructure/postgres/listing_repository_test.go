package postgres

import (
	"context"
	"testing"
	"time"

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
	wantMarket := off + " AND (km IS NULL OR km > 0)"
	if got := (&ListingRepository{}).marketWhere(); got != wantMarket {
		t.Fatalf("marketWhere = %q", got)
	}
}

func intPtr(v int) *int { return &v }

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

func TestMarketSummary_excludesZeroKm(t *testing.T) {
	pool := catalogPool(t)
	defer pool.Close()
	repo := NewListingRepository(pool, false)
	ctx := context.Background()
	trimID := fixtureTrim(t, pool)

	rows := []domain.Listing{
		{Source: "file", ExternalID: "km-used-1", TrimID: trimID, Year: 2077, KM: intPtr(40_000), Price: 20_000_000, Currency: "ARS"},
		{Source: "file", ExternalID: "km-used-2", TrimID: trimID, Year: 2077, KM: intPtr(55_000), Price: 22_000_000, Currency: "ARS"},
		{Source: "file", ExternalID: "km-new", TrimID: trimID, Year: 2077, KM: intPtr(0), Price: 80_000_000, Currency: "ARS"},
		{Source: "file", ExternalID: "km-null", TrimID: trimID, Year: 2077, Price: 21_000_000, Currency: "ARS"},
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM vehicle_listings WHERE external_id LIKE 'km-%'`)
	})
	for _, listing := range rows {
		if _, err := repo.Upsert(ctx, listing); err != nil {
			t.Fatal(err)
		}
	}

	summary, err := repo.MarketSummary(ctx, trimID, 2077, "ARS")
	if err != nil {
		t.Fatal(err)
	}
	if summary.Count != 3 {
		t.Fatalf("used count = %d", summary.Count)
	}
	if summary.Median == nil || *summary.Median != 21_000_000 {
		t.Fatalf("median = %v", summary.Median)
	}
	if summary.Maximum == nil || *summary.Maximum != 22_000_000 {
		t.Fatalf("max = %v", summary.Maximum)
	}

	list, err := repo.ListByTrimYear(ctx, trimID, 2077, 10, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 4 {
		t.Fatalf("list = %d", len(list))
	}
}

func TestExpireUnseen_sourceFilter(t *testing.T) {
	pool := catalogPool(t)
	defer pool.Close()
	repo := NewListingRepository(pool, false)
	ctx := context.Background()
	trimID := fixtureTrim(t, pool)

	ml := domain.Listing{Source: "zz-expire-ml", ExternalID: "exp-ml", TrimID: trimID, Year: 2019, Price: 1, Currency: "ARS"}
	file := domain.Listing{Source: "zz-expire-file", ExternalID: "exp-file", TrimID: trimID, Year: 2019, Price: 1, Currency: "ARS"}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM vehicle_listings WHERE external_id IN ('exp-ml', 'exp-file')`)
	})
	for _, listing := range []domain.Listing{ml, file} {
		if _, err := repo.Upsert(ctx, listing); err != nil {
			t.Fatal(err)
		}
	}
	_, err := pool.Exec(ctx, `
		UPDATE vehicle_listings
		SET last_seen_at = NOW() - INTERVAL '2 days'
		WHERE external_id IN ('exp-ml', 'exp-file')`)
	if err != nil {
		t.Fatal(err)
	}

	cutoff := time.Now().Add(-24 * time.Hour)
	n, err := repo.ExpireUnseen(ctx, cutoff, "zz-expire-ml")
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expired = %d", n)
	}

	var mlActive, fileActive bool
	if err := pool.QueryRow(ctx,
		`SELECT active FROM vehicle_listings WHERE external_id = 'exp-ml'`,
	).Scan(&mlActive); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx,
		`SELECT active FROM vehicle_listings WHERE external_id = 'exp-file'`,
	).Scan(&fileActive); err != nil {
		t.Fatal(err)
	}
	if mlActive || !fileActive {
		t.Fatalf("ml active=%v file active=%v", mlActive, fileActive)
	}
}

func fixtureNamedTrim(t *testing.T, pool *pgxpool.Pool, brand, model, trim string) int {
	t.Helper()
	ctx := context.Background()
	var id int
	err := pool.QueryRow(ctx, `
		WITH b AS (
			INSERT INTO brands (name) VALUES ($1)
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		), m AS (
			INSERT INTO models (brand_id, name)
			SELECT id, $2 FROM b
			ON CONFLICT (brand_id, name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		), g AS (
			INSERT INTO generations (model_id, name, year_from)
			SELECT id, $2, 2019 FROM m
			ON CONFLICT (model_id, name) DO UPDATE SET year_from = EXCLUDED.year_from
			RETURNING id
		)
		INSERT INTO trims (generation_id, name, year_from)
		SELECT id, $3, 2019 FROM g
		ON CONFLICT (generation_id, name) DO UPDATE SET year_from = EXCLUDED.year_from
		RETURNING id`, brand, model, trim).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM vehicle_listings WHERE trim_id = $1`, id)
		_, _ = pool.Exec(context.Background(), `DELETE FROM trims WHERE id = $1`, id)
		_, _ = pool.Exec(context.Background(), `DELETE FROM generations WHERE name = $1`, model)
		_, _ = pool.Exec(context.Background(), `DELETE FROM models WHERE name = $1`, model)
		_, _ = pool.Exec(context.Background(), `
			DELETE FROM brands WHERE name = $1 AND NOT EXISTS (
				SELECT 1 FROM models WHERE brand_id = brands.id
			)`, brand)
	})
	return id
}

func TestMarketSummary_observedAtAndSources(t *testing.T) {
	pool := catalogPool(t)
	defer pool.Close()
	repo := NewListingRepository(pool, false)
	ctx := context.Background()
	trimID := fixtureTrim(t, pool)

	older := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 10, 16, 16, 0, 0, 0, time.UTC)
	zeroKmSeen := time.Date(2026, 10, 17, 18, 0, 0, 0, time.UTC)
	rows := []domain.Listing{
		{Source: "mercadolibre", ExternalID: "obs-ml-1", TrimID: trimID, Year: 2066, KM: intPtr(40_000), Price: 20_000_000, Currency: "ARS", LastSeenAt: older},
		{Source: "mercadolibre", ExternalID: "obs-ml-2", TrimID: trimID, Year: 2066, KM: intPtr(55_000), Price: 22_000_000, Currency: "ARS", LastSeenAt: newer},
		{Source: "file", ExternalID: "obs-file", TrimID: trimID, Year: 2066, Price: 21_000_000, Currency: "ARS", LastSeenAt: older},
		{Source: "mercadolibre", ExternalID: "obs-new", TrimID: trimID, Year: 2066, KM: intPtr(0), Price: 80_000_000, Currency: "ARS", LastSeenAt: zeroKmSeen},
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM vehicle_listings WHERE external_id LIKE 'obs-%'`)
	})
	for _, listing := range rows {
		if _, err := repo.Upsert(ctx, listing); err != nil {
			t.Fatal(err)
		}
	}

	summary, err := repo.MarketSummary(ctx, trimID, 2066, "ARS")
	if err != nil {
		t.Fatal(err)
	}
	if summary.Count != 3 {
		t.Fatalf("count = %d", summary.Count)
	}
	if summary.ObservedAt == nil || !summary.ObservedAt.Equal(newer) {
		t.Fatalf("observed_at = %v", summary.ObservedAt)
	}
	if len(summary.Sources) != 2 {
		t.Fatalf("sources = %+v", summary.Sources)
	}
	if summary.Sources[0].Source != "mercadolibre" || summary.Sources[0].Count != 2 {
		t.Fatalf("first source = %+v", summary.Sources[0])
	}
	if summary.Sources[1].Source != "file" || summary.Sources[1].Count != 1 {
		t.Fatalf("second source = %+v", summary.Sources[1])
	}

	empty, err := repo.MarketSummary(ctx, trimID, 1901, "ARS")
	if err != nil {
		t.Fatal(err)
	}
	if empty.Count != 0 || empty.ObservedAt != nil || len(empty.Sources) != 0 {
		t.Fatalf("empty = %+v", empty)
	}
}

func TestListModelOffer_usedCatalogCounts(t *testing.T) {
	pool := catalogPool(t)
	defer pool.Close()
	repo := NewListingRepository(pool, true)
	ctx := context.Background()

	busy := fixtureNamedTrim(t, pool, "zz-offer-toyota", "zz-offer-corolla", "zz-offer-xei")
	thin := fixtureNamedTrim(t, pool, "zz-offer-fiat", "zz-offer-duna", "zz-offer-cl")
	onlyNew := fixtureNamedTrim(t, pool, "zz-offer-vw", "zz-offer-up", "zz-offer-take")

	seen := time.Date(2026, 10, 16, 15, 0, 0, 0, time.UTC)
	rows := []domain.Listing{
		{Source: "mercadolibre", ExternalID: "off-cor-1", TrimID: busy, Year: 2019, KM: intPtr(40_000), Price: 1, Currency: "ARS", LastSeenAt: seen},
		{Source: "mercadolibre", ExternalID: "off-cor-2", TrimID: busy, Year: 2020, KM: intPtr(30_000), Price: 1, Currency: "USD", LastSeenAt: seen},
		{Source: "mercadolibre", ExternalID: "off-duna", TrimID: thin, Year: 1995, KM: intPtr(120_000), Price: 1, Currency: "ARS", LastSeenAt: seen.Add(-time.Hour)},
		{Source: "mercadolibre", ExternalID: "off-up-new", TrimID: onlyNew, Year: 2024, KM: intPtr(0), Price: 1, Currency: "ARS", LastSeenAt: seen},
		{Source: "seed", ExternalID: "off-seed", TrimID: thin, Year: 1995, KM: intPtr(90_000), Price: 1, Currency: "ARS", LastSeenAt: seen},
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM vehicle_listings WHERE external_id LIKE 'off-%'`)
	})
	for _, listing := range rows {
		if _, err := repo.Upsert(ctx, listing); err != nil {
			t.Fatal(err)
		}
	}

	list, err := repo.ListModelOffer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]domain.ModelOffer{}
	for _, model := range list.Models {
		if len(model.ModelName) >= 9 && model.ModelName[:9] == "zz-offer-" {
			byName[model.ModelName] = model
		}
	}
	if byName["zz-offer-corolla"].Count != 2 {
		t.Fatalf("corolla = %+v", byName["zz-offer-corolla"])
	}
	if byName["zz-offer-duna"].Count != 1 {
		t.Fatalf("duna = %+v", byName["zz-offer-duna"])
	}
	if _, ok := byName["zz-offer-up"]; ok {
		t.Fatalf("0 km model in offer: %+v", byName["zz-offer-up"])
	}
	if list.ObservedAt == nil || !list.ObservedAt.Equal(seen) {
		t.Fatalf("observed_at = %v", list.ObservedAt)
	}

	detail, err := repo.GetModelOffer(ctx, byName["zz-offer-corolla"].ModelID)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Count != 2 || len(detail.Trims) != 1 || len(detail.Trims[0].Years) != 2 {
		t.Fatalf("detail = %+v", detail)
	}
	if _, err := repo.GetModelOffer(ctx, 0); err != domain.ErrNotFound {
		t.Fatalf("missing model err = %v", err)
	}
}
