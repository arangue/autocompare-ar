package postgres

import (
	"context"
	"testing"

	"github.com/arangue/autocompare-ar/internal/domain"
)

func TestUpsert_activeFalsePersists(t *testing.T) {
	pool := catalogPool(t)
	defer pool.Close()
	repo := NewListingRepository(pool)
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
	repo := NewListingRepository(pool)
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
