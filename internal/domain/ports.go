package domain

import "context"

// Ports (driven): interfaces the domain/application layer depends on.
// Infrastructure adapters implement these.

type VehicleRepository interface {
	ListBrands(ctx context.Context) ([]Brand, error)
	ListModelsByBrand(ctx context.Context, brandID int) ([]Model, error)
	SearchTrims(ctx context.Context, query string, limit int) ([]TrimSearchResult, error)
	GetTrim(ctx context.Context, trimID int) (TrimDetail, error)
}

type ListingRepository interface {
	ListByTrimYear(ctx context.Context, trimID, year, limit int) ([]Listing, error)
	MarketSummary(ctx context.Context, trimID, year int) (MarketSummary, error)
	// Upsert inserts or updates by (source, external_id). inserted is true on insert.
	Upsert(ctx context.Context, listing Listing) (inserted bool, err error)
	// ExpireStale sets active=false on listings not seen in the last days.
	// days <= 0 is a no-op. Does not DELETE rows.
	ExpireStale(ctx context.Context, days int) (int64, error)
}

type PricingRepository interface {
	ListReferences(ctx context.Context, trimID, year int) ([]ReferencePrice, error)
}

type AliasRepository interface {
	Resolve(ctx context.Context, source, raw string) (trimID int, ok bool, err error)
}
