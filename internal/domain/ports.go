package domain

import "context"

// Ports (driven): interfaces the domain/application layer depends on.
// Infrastructure adapters implement these.

type VehicleRepository interface {
	ListBrands(ctx context.Context) ([]Brand, error)
	ListModelsByBrand(ctx context.Context, brandID int) ([]Model, error)
	SearchTrims(ctx context.Context, query string, limit int) ([]TrimSearchResult, error)
}

type ListingRepository interface {
	ListByTrimYear(ctx context.Context, trimID, year, limit int) ([]Listing, error)
	MarketSummary(ctx context.Context, trimID, year int) (MarketSummary, error)
}

type PricingRepository interface {
	ListReferences(ctx context.Context, trimID, year int) ([]ReferencePrice, error)
}
