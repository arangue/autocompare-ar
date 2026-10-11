package ingestion_test

import (
	"testing"

	"github.com/arangue/autocompare-ar/internal/infrastructure/ingestion"
)

func TestParserRegistry_listingsOnly(t *testing.T) {
	registry := ingestion.NewDefaultRegistry()

	if _, err := registry.ParserFor("unknown.xyz"); err == nil {
		t.Fatal("expected error for unknown file extension")
	}
	if _, err := registry.ParserFor("document.pdf"); err == nil {
		t.Fatal("pdf is a price guide, not a listing file")
	}
	if _, err := registry.ParserFor("listings.json"); err != nil {
		t.Fatal(err)
	}
}
