package postgres

import (
	"context"
	"testing"

	"github.com/arangue/autocompare-ar/internal/domain"
)

func TestUpsert_activeFalsePersists(t *testing.T) {
	pool := catalogPool(t)
	defer pool.Close()
	repo := NewListingRepository(pool, false)
	ctx := context.Background()

	off := false
	listing := domain.Listing{
		Source:     "file",
		ExternalID: "e4-402-inactive",
		TrimID:     1,
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

	got, err := repo.ListByTrimYear(ctx, listing.TrimID, listing.Year, 100)
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

	listing := domain.Listing{
		Source:     "file",
		ExternalID: "e4-403-stale",
		TrimID:     1,
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

	got, err := repo.ListByTrimYear(ctx, listing.TrimID, listing.Year, 100)
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

	seed := domain.Listing{Source: "seed", ExternalID: "e5-502-seed", TrimID: 1, Year: 2099, Price: 100, Currency: "ARS"}
	live := domain.Listing{Source: "file", ExternalID: "e5-502-file", TrimID: 1, Year: 2099, Price: 200, Currency: "ARS"}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM vehicle_listings WHERE external_id IN ('e5-502-seed', 'e5-502-file')`)
	})
	for _, listing := range []domain.Listing{seed, live} {
		if _, err := off.Upsert(ctx, listing); err != nil {
			t.Fatal(err)
		}
	}

	offMarket, err := off.MarketSummary(ctx, 1, 2099)
	if err != nil {
		t.Fatal(err)
	}
	offList, err := off.ListByTrimYear(ctx, 1, 2099, 100)
	if err != nil {
		t.Fatal(err)
	}
	if offMarket.Count != 2 || len(offList) != 2 {
		t.Fatalf("flag off market=%d list=%d", offMarket.Count, len(offList))
	}

	onMarket, err := on.MarketSummary(ctx, 1, 2099)
	if err != nil {
		t.Fatal(err)
	}
	onList, err := on.ListByTrimYear(ctx, 1, 2099, 100)
	if err != nil {
		t.Fatal(err)
	}
	if onMarket.Count != 1 || len(onList) != 1 || onList[0].Source != "file" {
		t.Fatalf("flag on market=%d list=%v", onMarket.Count, onList)
	}
}
