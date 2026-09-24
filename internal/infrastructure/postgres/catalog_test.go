package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// year_from/year_to NULL is not inverted; empty name is.
const catalogBadTrimSQL = `
	SELECT COUNT(*) FROM trims
	WHERE name = ''
	   OR (year_from IS NOT NULL AND year_to IS NOT NULL AND year_from > year_to)`

func TestCatalog_seed(t *testing.T) {
	pool := catalogPool(t)
	defer pool.Close()

	n := catalogBadCount(t, context.Background(), pool)
	if n != 0 {
		t.Fatalf("seed has %d invalid trims (empty name or year_from > year_to)", n)
	}
}

func TestCatalog_rejectsBrokenFixture(t *testing.T) {
	pool := catalogPool(t)
	defer pool.Close()

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO trims (generation_id, name, year_from, year_to)
		SELECT id, 'E3-301-bad', 2020, 2018 FROM generations LIMIT 1`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO trims (generation_id, name, year_from, year_to)
		SELECT id, '', 2019, 2020 FROM generations LIMIT 1`)
	if err != nil {
		t.Fatal(err)
	}

	n := catalogBadCount(t, ctx, tx)
	if n < 2 {
		t.Fatalf("broken fixture counted %d, want ≥2", n)
	}
}

func catalogPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Fatal("DATABASE_URL is not set; make db-up && make api first")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("catalog-check requires Postgres (make db-up): %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("catalog-check requires Postgres (make db-up): %v", err)
	}
	return pool
}

type catalogQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func catalogBadCount(t *testing.T, ctx context.Context, q catalogQuerier) int {
	t.Helper()
	var n int
	if err := q.QueryRow(ctx, catalogBadTrimSQL).Scan(&n); err != nil {
		t.Fatalf("query trims: %v (run make api so migrations are applied)", err)
	}
	return n
}
