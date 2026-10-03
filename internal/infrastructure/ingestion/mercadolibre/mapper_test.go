package mercadolibre

import "testing"

func TestMapItem(t *testing.T) {
	km := 101000
	loc := "Rosario, Santa Fe"
	url := "https://auto.mercadolibre.com.ar/MLA-123"
	got, ok := MapItem(rawItem{
		ID:         "MLA123",
		Title:      "Toyota Corolla XEI 2.0 CVT",
		Price:      24800000,
		CurrencyID: "ARS",
		Permalink:  url,
		Address:    &place{CityName: "Rosario", StateName: "Santa Fe"},
		Attributes: []attribute{
			{ID: "VEHICLE_YEAR", ValueName: "2019"},
			{ID: "KILOMETERS", ValueName: "101.000 km"},
			{ID: "TRIM", ValueName: "XEi 2.0 CVT"},
		},
	})
	if !ok {
		t.Fatal("expected ok")
	}
	if got.Source != "mercadolibre" || got.ExternalID != "MLA123" || got.RawTitle != "Toyota Corolla XEI 2.0 CVT" {
		t.Fatalf("identity = %+v", got)
	}
	if got.TrimID != 0 {
		t.Fatalf("trim_id = %d", got.TrimID)
	}
	if got.Year != 2019 || got.Price != 24800000 || got.Currency != "ARS" {
		t.Fatalf("numbers = %+v", got)
	}
	if got.KM == nil || *got.KM != km {
		t.Fatalf("km = %v", got.KM)
	}
	if got.Location == nil || *got.Location != loc {
		t.Fatalf("location = %v", got.Location)
	}
	if got.URL == nil || *got.URL != url {
		t.Fatalf("url = %v", got.URL)
	}
}

func TestMapItem_skips(t *testing.T) {
	base := rawItem{
		ID:    "MLA1",
		Title: "Toyota Corolla",
		Price: 1,
		Attributes: []attribute{
			{ID: "VEHICLE_YEAR", ValueName: "2019"},
		},
	}
	cases := []rawItem{
		{Title: "x", Price: 1, Attributes: base.Attributes},
		{ID: "MLA1", Price: 1, Attributes: base.Attributes},
		{ID: "MLA1", Title: "x", Attributes: base.Attributes},
		{ID: "MLA1", Title: "x", Price: 1},
	}
	for i, item := range cases {
		if _, ok := MapItem(item); ok {
			t.Fatalf("case %d: expected skip", i)
		}
	}
}

func TestMapItem_kmOptional(t *testing.T) {
	got, ok := MapItem(rawItem{
		ID:    "MLA9",
		Title: "Golf",
		Price: 10,
		Attributes: []attribute{
			{ID: "VEHICLE_YEAR", ValueName: "2018"},
		},
	})
	if !ok {
		t.Fatal("expected ok")
	}
	if got.KM != nil {
		t.Fatalf("km = %v", got.KM)
	}
	if got.Currency != "ARS" {
		t.Fatalf("currency = %q", got.Currency)
	}
}
