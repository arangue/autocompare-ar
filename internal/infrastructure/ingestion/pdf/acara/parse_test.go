package acara

import (
	"os"
	"strconv"
	"testing"
	"time"
)

func TestParseText_guideColumns(t *testing.T) {
	const sample = `Guía Oficial de Precios TOYOTA 0km 2025 2024 2023 2022 2021 2020 2019 2018 2017 2016 2015 2014 2013 2012 2011 Corolla 2.0 XEI CVT (170cv) (L23) $ 48.410.000 39.264.100 34.159.800 29.719.000 - - - - - - - - - - - - 2.0 XLI CVT (170cv) (L23) $ 44.082.000 35.788.700 31.136.200 27.088.500 - - - - - - - - - - - - Guía Oficial de Precios VOLKSWAGEN 0km 2025 2024 2023 2022 2021 2020 2019 2018 2017 2016 2015 2014 2013 2012 2011 Golf 1.4 TSI Comfortline MT (150cv) (L17) $ - - - - - - - 25.235.500 22.844.900 20.903.900 - - - - - - Guía Oficial de Precios FIAT 0km 2025 2024 2023 2022 2021 2020 2019 2018 2017 2016 2015 2014 2013 2012 2011 Cronos 1.3 GSE Drive MT (99cv) $ - - - - - 19.338.200 16.701.300 15.822.200 14.943.100 - - - - - - - 1.3 GSE Drive MT (99cv) (L23) $ - - - 21.621.600 21.096.300 - - - - - - - - - - - `
	observed := time.Date(2026, 9, 29, 0, 0, 0, 0, time.UTC)
	rows, err := ParseText(sample, observed)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 5 {
		t.Fatalf("rows = %d, want 5", len(rows))
	}

	xei := rows[0]
	if xei.Brand != "TOYOTA" || xei.Model != "Corolla" || xei.Currency != "ARS" {
		t.Fatalf("xei row = %+v", xei)
	}
	if got := xei.Keys()[0]; got != "TOYOTA Corolla 2.0 XEI CVT" {
		t.Fatalf("key = %q", got)
	}
	assertCell(t, xei, 2026, 48410000)
	assertCell(t, xei, 2023, 29719000)
	if rows[1].Keys()[0] != "TOYOTA Corolla 2.0 XLI CVT" {
		t.Fatalf("continuation key = %q", rows[1].Keys()[0])
	}
	if _, ok := cell(xei, 2019); ok {
		t.Fatal("L23 XEi has no 2019 column")
	}

	golf := rows[2]
	if golf.Keys()[0] != "VOLKSWAGEN Golf 1.4 TSI Comfortline MT" {
		t.Fatalf("golf key = %q", golf.Keys()[0])
	}
	assertCell(t, golf, 2019, 25235500)

	cronos := rows[3]
	if cronos.Keys()[0] != "FIAT Cronos 1.3 GSE Drive MT" {
		t.Fatalf("cronos key = %q", cronos.Keys()[0])
	}
	assertCell(t, cronos, 2019, 15822200)
}

func TestParseAmount_fullUnits(t *testing.T) {
	n, ok := parseAmount("48.410.000")
	if !ok || n != 48410000 {
		t.Fatalf("pesos = %d ok=%v", n, ok)
	}
	n, ok = parseAmount("60.900,00")
	if !ok || n != 60900 {
		t.Fatalf("usd = %d ok=%v", n, ok)
	}
	if _, ok := parseAmount("-"); ok {
		t.Fatal("dash should be empty")
	}
}

func TestParseFile_optionalGuide(t *testing.T) {
	path := os.Getenv("ACARA_PDF")
	if path == "" {
		t.Skip("set ACARA_PDF to check a real guide")
	}
	observed := time.Date(2026, 9, 29, 0, 0, 0, 0, time.UTC)
	rows, err := ParseFile(path, observed)
	if err != nil {
		t.Fatal(err)
	}
	var xei, golf, cronos *Row
	for i := range rows {
		r := &rows[i]
		switch r.Keys()[0] {
		case "TOYOTA Corolla 2.0 XEI CVT":
			if xei == nil {
				xei = r
			}
		case "VOLKSWAGEN Golf 1.4 TSI Comfortline MT":
			if golf == nil {
				golf = r
			}
		case "FIAT Cronos 1.3 GSE Drive MT":
			if cronos == nil {
				cronos = r
			}
		}
	}
	if xei == nil || golf == nil || cronos == nil {
		t.Fatalf("missing rows xei=%v golf=%v cronos=%v (parsed %d)", xei != nil, golf != nil, cronos != nil, len(rows))
	}
	assertCell(t, *xei, 2026, 48410000)
	assertCell(t, *xei, 2023, 29719000)
	assertCell(t, *golf, 2019, 25235500)
	assertCell(t, *cronos, 2019, 15822200)
	seen := map[string]int64{}
	for i := range rows {
		key := rows[i].Keys()[0]
		if key != "TOYOTA Corolla 2.0 XEI CVT" && key != "FIAT Cronos 1.3 GSE Drive MT" && key != "VOLKSWAGEN Golf 1.4 TSI Comfortline MT" {
			continue
		}
		for _, c := range rows[i].Cells {
			id := key + "/" + strconv.Itoa(c.Year)
			if prev, ok := seen[id]; ok && prev != c.Price {
				t.Errorf("overlap %s %d then %d", id, prev, c.Price)
			}
			seen[id] = c.Price
		}
	}
}

func assertCell(t *testing.T, row Row, year int, price int64) {
	t.Helper()
	got, ok := cell(row, year)
	if !ok || got != price {
		t.Fatalf("%s %s %d = %d ok=%v, want %d", row.Brand, row.Model, year, got, ok, price)
	}
}

func cell(row Row, year int) (int64, bool) {
	for _, c := range row.Cells {
		if c.Year == year {
			return c.Price, true
		}
	}
	return 0, false
}
