package mercadolibre

import "testing"

func TestParseSearchCSV(t *testing.T) {
	raw := `Title,Price (ARS),Year,Kilometers,Location,URL
Ford Focus S 1.6 MT 5P,16.500.000,2019,134.000 Km,Hurlingham,https://auto.mercadolibre.com.ar/MLA-2168520435-ford-focus-_JM
Ford Ka,128.893.609,2020,0 Km,Buenos Aires,https://auto.mercadolibre.com.ar/MLA-9-ka-_JM
Bad,N/A,2019,10 Km,X,https://auto.mercadolibre.com.ar/MLA-1
NoID,100,2019,10 Km,X,https://example.com/nope
NoYear,100,,10 Km,X,https://auto.mercadolibre.com.ar/MLA-2
`
	listings, skipped, err := ParseSearchCSV([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if skipped != 3 {
		t.Fatalf("skipped = %d", skipped)
	}
	if len(listings) != 2 {
		t.Fatalf("listings = %d", len(listings))
	}

	got := listings[0]
	if got.Source != "mercadolibre" || got.ExternalID != "MLA2168520435" || got.TrimID != 0 {
		t.Fatalf("identity = %+v", got)
	}
	if got.RawTitle != "Ford Focus S 1.6 MT 5P" || got.Year != 2019 || got.Price != 16500000 || got.Currency != "ARS" {
		t.Fatalf("numbers = %+v", got)
	}
	if got.KM == nil || *got.KM != 134000 {
		t.Fatalf("km = %v", got.KM)
	}
	if got.Location == nil || *got.Location != "Hurlingham" {
		t.Fatalf("location = %v", got.Location)
	}
	if got.URL == nil || *got.URL != "https://auto.mercadolibre.com.ar/MLA-2168520435-ford-focus-_JM" {
		t.Fatalf("url = %v", got.URL)
	}

	if listings[1].ExternalID != "MLA9" || listings[1].Price != 128893609 {
		t.Fatalf("second = %+v", listings[1])
	}
	if listings[1].KM == nil || *listings[1].KM != 0 {
		t.Fatalf("zero km = %v", listings[1].KM)
	}
}

func TestParseSearchCSV_currencyColumn(t *testing.T) {
	raw := `Title,Price,Currency,Year,Kilometers,Location,URL
Honda Civic,48.000,USD,2026,0 Km,Tucumán,https://auto.mercadolibre.com.ar/MLA-1-civic-_JM
Honda CR-V,89.900.000,ARS,2025,0 Km,CABA,https://auto.mercadolibre.com.ar/MLA-2-crv-_JM
Honda HR-V,19.800,US$,2019,90.000 Km,CABA,https://auto.mercadolibre.com.ar/MLA-3-hrv-_JM
`
	listings, skipped, err := ParseSearchCSV([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if skipped != 0 || len(listings) != 3 {
		t.Fatalf("listings = %d skipped = %d", len(listings), skipped)
	}
	if listings[0].Price != 48000 || listings[0].Currency != "USD" {
		t.Fatalf("usd = %+v", listings[0])
	}
	if listings[1].Price != 89900000 || listings[1].Currency != "ARS" {
		t.Fatalf("ars = %+v", listings[1])
	}
	if listings[2].Price != 19800 || listings[2].Currency != "USD" {
		t.Fatalf("symbol = %+v", listings[2])
	}
}

func TestParseSearchCSV_missingColumn(t *testing.T) {
	_, _, err := ParseSearchCSV([]byte("Title,Year\nFocus,2019\n"))
	if err == nil {
		t.Fatal("expected missing column error")
	}
}

func TestExternalID(t *testing.T) {
	id, ok := ExternalID("https://auto.mercadolibre.com.ar/MLA-2168520435-ford-focus-_JM")
	if !ok || id != "MLA2168520435" {
		t.Fatalf("id = %q %v", id, ok)
	}
	if _, ok := ExternalID("https://example.com/nope"); ok {
		t.Fatal("expected no id")
	}
}
