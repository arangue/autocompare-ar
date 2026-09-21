package ingestion

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile_JSON(t *testing.T) {
	path := writeTemp(t, "listings.json", `[
		{"source":"file","external_id":"a","trim_id":1,"year":2019,"km":98000,"price":24800000,"currency":"ARS","location":"Rosario","url":"https://example.com/a"},
		{"source":"file","external_id":"bad","trim_id":0,"year":2019,"price":1}
	]`)

	listings, skipped, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if skipped != 1 {
		t.Fatalf("skipped = %d", skipped)
	}
	if len(listings) != 1 {
		t.Fatalf("len = %d", len(listings))
	}
	got := listings[0]
	if got.Source != "file" || got.ExternalID != "a" || got.TrimID != 1 || got.Year != 2019 || got.Price != 24800000 {
		t.Fatalf("listing = %+v", got)
	}
	if got.KM == nil || *got.KM != 98000 {
		t.Fatalf("km = %v", got.KM)
	}
	if got.Currency != "ARS" {
		t.Fatalf("currency = %q", got.Currency)
	}
}

func TestParseFile_CSV(t *testing.T) {
	path := writeTemp(t, "listings.csv", "source,external_id,trim_id,year,km,price,currency,location,url\n"+
		"file,b,1,2019,76000,25800000,,Córdoba,https://example.com/b\n"+
		"file,bad,1,0,1,0,,,\n")

	listings, skipped, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if skipped != 1 || len(listings) != 1 {
		t.Fatalf("len=%d skipped=%d", len(listings), skipped)
	}
	if listings[0].Currency != "ARS" {
		t.Fatalf("default currency = %q", listings[0].Currency)
	}
	if listings[0].ExternalID != "b" || listings[0].Price != 25800000 {
		t.Fatalf("listing = %+v", listings[0])
	}
}

func TestParseJSON_rawTitleWithoutTrimID(t *testing.T) {
	path := writeTemp(t, "listings.json", `[
	  {"source":"file","external_id":"x","year":2019,"price":20000000,
	   "raw_title":"Toyota Corolla XEI 2.0 CVT"}
	]`)

	listings, skipped, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if skipped != 0 || len(listings) != 1 {
		t.Fatalf("len=%d skipped=%d", len(listings), skipped)
	}
	got := listings[0]
	if got.TrimID != 0 {
		t.Fatalf("trim_id = %d", got.TrimID)
	}
	if got.RawTitle != "Toyota Corolla XEI 2.0 CVT" {
		t.Fatalf("raw_title = %q", got.RawTitle)
	}
}

func TestParseFile_unsupported(t *testing.T) {
	path := writeTemp(t, "listings.txt", "nope")
	if _, _, err := ParseFile(path); err == nil {
		t.Fatal("expected error")
	}
}

func writeTemp(t *testing.T, name, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
