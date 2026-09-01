package main

import (
	"testing"

	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/arangue/autocompare-ar/migrations"
)

func TestMigrateDatabaseURL(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{
			"postgres://user:pass@localhost:5435/autocompare_db?sslmode=disable",
			"pgx5://user:pass@localhost:5435/autocompare_db?sslmode=disable",
		},
		{
			"postgresql://user:pass@db.example:5432/autocompare_db?sslmode=require",
			"pgx5://user:pass@db.example:5432/autocompare_db?sslmode=require",
		},
		{
			"pgx5://already-converted",
			"pgx5://already-converted",
		},
	}
	for _, tt := range tests {
		if got := migrateDatabaseURL(tt.in); got != tt.want {
			t.Fatalf("migrateDatabaseURL(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestEmbeddedMigrationsOpen(t *testing.T) {
	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		t.Fatal(err)
	}
	version, err := d.First()
	if err != nil {
		t.Fatal(err)
	}
	if version != 1 {
		t.Fatalf("first version = %d, want 1", version)
	}
}
