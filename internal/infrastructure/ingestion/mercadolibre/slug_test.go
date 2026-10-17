package mercadolibre

import "testing"

func TestSlug(t *testing.T) {
	cases := map[string]string{
		"Citroën":       "citroen",
		"Mercedes-Benz": "mercedes-benz",
		"Grand Siena":   "grand-siena",
		"Clase C":       "clase-c",
		"Tiggo 7 Pro":   "tiggo-7-pro",
		"ford":          "ford",
		"  Fiat  ":      "fiat",
	}
	for in, want := range cases {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSearchPath(t *testing.T) {
	got := SearchPath("Mercedes-Benz", "Clase C")
	if got != "mercedes-benz-clase-c" {
		t.Fatalf("got %q", got)
	}
	got = SearchPath("Fiat", "Siena")
	if got != "fiat-siena" {
		t.Fatalf("got %q", got)
	}
	got = SearchPath("Citroën", "C3")
	if got != "citroen-c3" {
		t.Fatalf("got %q", got)
	}
}
