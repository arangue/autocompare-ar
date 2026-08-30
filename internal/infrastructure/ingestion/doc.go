// Package ingestion imports classified listings into vehicle_listings via
// ListingRepository.Upsert. file_listings.go reads local JSON/CSV. Guide prices
// (CCA/ACARA/DNRPA) are not ingested here — see DESIGN.md §9.
package ingestion
