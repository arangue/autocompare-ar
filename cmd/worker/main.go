package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arangue/autocompare-ar/internal/domain"
	"github.com/arangue/autocompare-ar/internal/infrastructure/ingestion"
	"github.com/arangue/autocompare-ar/internal/infrastructure/ingestion/carsbase"
	"github.com/arangue/autocompare-ar/internal/infrastructure/ingestion/normalize"
	"github.com/arangue/autocompare-ar/internal/infrastructure/ingestion/pdf/acara"
	"github.com/arangue/autocompare-ar/internal/infrastructure/ingestion/pdf/cca"
	"github.com/arangue/autocompare-ar/internal/infrastructure/ingestion/pdf/guide"
	"github.com/arangue/autocompare-ar/internal/infrastructure/postgres"
	"github.com/arangue/autocompare-ar/migrations"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dryRun := flag.Bool("dry-run", false, "log matches without writing")
	expireDays := flag.Int("expire-days", 0, "deactivate listings not seen in N days")
	source := flag.String("source", "", "acara, cca, or carsbase")
	observed := flag.String("observed", "", "guide date YYYY-MM-DD (default today)")
	flag.Parse()

	path := flag.Arg(0)
	if path == "" {
		path = os.Getenv("INGEST_FILE")
	}
	if *source == "acara" || *source == "cca" || *source == "carsbase" {
		if path == "" {
			slog.Error("provide a file path")
			os.Exit(1)
		}
		if *source == "carsbase" {
			if err := runCarsbase(path, *dryRun); err != nil {
				slog.Error("carsbase ingest failed", "error", err)
				os.Exit(1)
			}
			return
		}
		if err := runGuide(path, *observed, *dryRun, *source); err != nil {
			slog.Error("guide ingest failed", "error", err)
			os.Exit(1)
		}
		return
	}
	if *source != "" {
		slog.Error("unknown source", "source", *source)
		os.Exit(1)
	}
	if path == "" && *expireDays <= 0 {
		slog.Error("provide a file path argument or INGEST_FILE, or -expire-days N")
		os.Exit(1)
	}

	var listings []domain.Listing
	skipped := 0
	if path != "" {
		var err error
		listings, skipped, err = ingestion.ParseFile(path)
		if err != nil {
			slog.Error("failed to load listings", "error", err)
			os.Exit(1)
		}
	}

	needResolve := false
	for _, listing := range listings {
		if listing.TrimID <= 0 {
			needResolve = true
			break
		}
	}

	needDB := !*dryRun || needResolve
	var aliasRepo domain.AliasRepository
	var listingRepo *postgres.ListingRepository
	ctx := context.Background()
	if needDB {
		dbURL := databaseURL()

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
		listingRepo = postgres.NewListingRepository(pool, false)
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

	expired := int64(0)
	if !*dryRun && *expireDays > 0 {
		var err error
		expired, err = listingRepo.ExpireStale(ctx, *expireDays)
		if err != nil {
			slog.Error("expire failed", "error", err)
			os.Exit(1)
		}
	}

	slog.Info("ingest finished",
		"dry_run", *dryRun,
		"matched", len(listings),
		"inserted", inserted,
		"updated", updated,
		"skipped", skipped,
		"expired", expired,
	)
}

func runGuide(path, observedRaw string, dryRun bool, kind string) error {
	observed := time.Now().UTC()
	if observedRaw != "" {
		t, err := time.Parse("2006-01-02", observedRaw)
		if err != nil {
			return fmt.Errorf("observed: %w", err)
		}
		observed = t.UTC()
	}
	var (
		rows        []guide.Row
		err         error
		refSource   string
		aliasSource string
	)
	switch kind {
	case "acara":
		rows, err = acara.ParseFile(path, observed)
		refSource, aliasSource = "ACARA", "acara"
	case "cca":
		rows, err = cca.ParseFile(path, observed)
		refSource, aliasSource = "CCA", "cca"
	default:
		return fmt.Errorf("unknown guide source %s", kind)
	}
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	if !dryRun {
		if err := applyMigrations(databaseURL()); err != nil {
			return err
		}
	}
	pool, err := openPool(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	aliases := postgres.NewAliasRepository(pool)
	prices := postgres.NewPricingRepository(pool)
	cache := map[string]resolveHit{}
	matched, stored, linked := 0, 0, 0
	for _, row := range rows {
		trimID, ok, err := resolveGuide(ctx, aliases, cache, row, aliasSource)
		if err != nil {
			return err
		}
		if ok {
			matched++
		}
		for _, cell := range row.Cells {
			if dryRun {
				continue
			}
			if _, err := prices.UpsertGuidePrice(ctx, refSource, row.Brand, row.Model, row.Trim, row.Currency, cell.Year, cell.Price, observed); err != nil {
				return fmt.Errorf("guide price %s %s %d: %w", row.Brand, row.Trim, cell.Year, err)
			}
			stored++
			if !ok {
				continue
			}
			if _, err := prices.UpsertReference(ctx, domain.ReferencePrice{
				TrimID:     trimID,
				Year:       cell.Year,
				Source:     refSource,
				Kind:       "guide",
				Price:      cell.Price,
				Currency:   row.Currency,
				ObservedAt: observed,
			}); err != nil {
				return fmt.Errorf("upsert %s %s %d: %w", row.Brand, row.Trim, cell.Year, err)
			}
			linked++
		}
	}
	slog.Info("guide ingest finished",
		"source", refSource,
		"dry_run", dryRun,
		"rows", len(rows),
		"matched", matched,
		"stored", stored,
		"linked", linked,
	)
	return nil
}

type resolveHit struct {
	id int
	ok bool
}

func resolveGuide(ctx context.Context, aliases domain.AliasRepository, cache map[string]resolveHit, row guide.Row, aliasSource string) (int, bool, error) {
	for _, key := range row.Keys() {
		norm := normalize.Title(key)
		for _, source := range []string{aliasSource, "file"} {
			cacheKey := source + "\x00" + norm
			if hit, ok := cache[cacheKey]; ok {
				if hit.ok {
					return hit.id, true, nil
				}
				continue
			}
			id, ok, err := aliases.Resolve(ctx, source, norm)
			if err != nil {
				return 0, false, err
			}
			cache[cacheKey] = resolveHit{id: id, ok: ok}
			if ok {
				return id, true, nil
			}
		}
	}
	return 0, false, nil
}

func runCarsbase(path string, dryRun bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	if !dryRun {
		if err := applyMigrations(databaseURL()); err != nil {
			return err
		}
	}
	pool, err := openPool(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	stats, err := carsbase.Load(ctx, pool, path, dryRun)
	if err != nil {
		return err
	}
	slog.Info("carsbase ingest finished",
		"dry_run", dryRun,
		"sheets", stats.Sheets,
		"dropped", stats.Dropped,
		"trims", stats.Trims,
		"links", stats.Links,
	)
	return nil
}

func applyMigrations(dbURL string) error {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", source, migrateDatabaseURL(dbURL))
	if err != nil {
		return err
	}
	defer m.Close() //nolint:errcheck
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func migrateDatabaseURL(dbURL string) string {
	switch {
	case strings.HasPrefix(dbURL, "postgresql://"):
		return "pgx5://" + strings.TrimPrefix(dbURL, "postgresql://")
	case strings.HasPrefix(dbURL, "postgres://"):
		return "pgx5://" + strings.TrimPrefix(dbURL, "postgres://")
	default:
		return dbURL
	}
}

func openPool(ctx context.Context) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL())
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

// Same as Makefile: .env if present, else docker-compose default.
func databaseURL() string {
	loadDotEnv(".env")
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://user:password@localhost:5435/autocompare_db?sslmode=disable"
}

func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		if k == "" || os.Getenv(k) != "" {
			continue
		}
		_ = os.Setenv(k, v)
	}
}
