package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arangue/autocompare-ar/internal/domain"
	"github.com/arangue/autocompare-ar/internal/infrastructure/ingestion"
	"github.com/arangue/autocompare-ar/internal/infrastructure/ingestion/normalize"
	"github.com/arangue/autocompare-ar/internal/infrastructure/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dryRun := flag.Bool("dry-run", false, "log matches without writing")
	flag.Parse()

	listings, skipped, err := loadListings()
	if err != nil {
		slog.Error("failed to load listings", "error", err)
		os.Exit(1)
	}

	needResolve := false
	for _, listing := range listings {
		if listing.TrimID <= 0 {
			needResolve = true
			break
		}
	}

	var aliasRepo domain.AliasRepository
	var listingRepo *postgres.ListingRepository
	ctx := context.Background()
	if !*dryRun || needResolve {
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			slog.Error("DATABASE_URL is not set")
			os.Exit(1)
		}

		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		pool, err := pgxpool.New(ctx, dbURL)
		if err != nil {
			slog.Error("failed to connect to database", "error", err)
			os.Exit(1)
		}
		defer pool.Close()

		if err := pool.Ping(ctx); err != nil {
			slog.Error("database is unreachable", "error", err)
			os.Exit(1)
		}

		aliasRepo = postgres.NewAliasRepository(pool)
		listingRepo = postgres.NewListingRepository(pool)
	}

	inserted, updated := 0, 0
	for _, listing := range listings {
		if listing.TrimID <= 0 {
			trimID, ok, err := aliasRepo.Resolve(ctx, listing.Source, normalize.Title(listing.RawTitle))
			if err != nil {
				slog.Error("resolve failed", "source", listing.Source, "external_id", listing.ExternalID, "error", err)
				os.Exit(1)
			}
			if !ok {
				slog.Warn("skipped listing", "source", listing.Source, "external_id", listing.ExternalID, "raw_title", listing.RawTitle)
				skipped++
				continue
			}
			listing.TrimID = trimID
		}

		if *dryRun {
			slog.Info("dry-run match",
				"source", listing.Source,
				"external_id", listing.ExternalID,
				"trim_id", listing.TrimID,
				"year", listing.Year,
				"price", listing.Price,
			)
			continue
		}

		wasInsert, err := listingRepo.Upsert(ctx, listing)
		if err != nil {
			if errors.Is(err, postgres.ErrUnknownTrim) {
				slog.Warn("skipped listing", "source", listing.Source, "external_id", listing.ExternalID, "error", err)
				skipped++
				continue
			}
			slog.Error("upsert failed", "source", listing.Source, "external_id", listing.ExternalID, "error", err)
			os.Exit(1)
		}
		if wasInsert {
			inserted++
		} else {
			updated++
		}
	}

	slog.Info("ingest finished",
		"dry_run", *dryRun,
		"matched", len(listings),
		"inserted", inserted,
		"updated", updated,
		"skipped", skipped,
	)
}

func loadListings() ([]domain.Listing, int, error) {
	path := flag.Arg(0)
	if path == "" {
		path = os.Getenv("INGEST_FILE")
	}
	if path == "" {
		return nil, 0, errors.New("provide a file path argument or INGEST_FILE")
	}
	return ingestion.ParseFile(path)
}
