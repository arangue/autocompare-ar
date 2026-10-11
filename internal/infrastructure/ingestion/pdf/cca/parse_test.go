package cca

import (
	"os"
	"testing"
	"time"

	"github.com/arangue/autocompare-ar/internal/infrastructure/ingestion/pdf/guide"
)

func TestParsePrice(t *testing.T) {
	n, ok := parsePrice("111,660")
	if !ok || n != 111660 {
		t.Fatalf("thousands = %d ok=%v", n, ok)
	}
	n, ok = parsePrice("48410,0")
	if !ok || n != 48410 {
		t.Fatalf("fraction = %d ok=%v", n, ok)
	}
	n, ok = parsePrice("13,346")
	if !ok || n != 13346 {
		t.Fatalf("grouped = %d ok=%v", n, ok)
	}
	if _, ok := parsePrice("-"); ok {
		t.Fatal("dash should be empty")
	}
}

func TestParsePDF_grid(t *testing.T) {
	const cmapBold = "begincmap\n1 beginbfrange\n<01><01><0041>\n<20><7e><0020>\nendbfrange\nendcmap\n"
	const cmapItalic = "begincmap\n1 beginbfrange\n<01><01><004d>\n<20><7e><0020>\nendbfrange\nendcmap\n"
	const content = "" +
		"615 128 m\n615 200 l\n660 200 l\n660 128 l\nh\nW n\n" +
		"q\n4.16667 0 0 4.16667 0 0 cm BT\n/R10 10 Tf\n[(FIAT)]TJ\nET\n" +
		"667 128 m\n667 200 l\n710 200 l\n710 128 l\nh\nW n\n" +
		"q\n4.16667 0 0 4.16667 0 0 cm BT\n/R12 10 Tf\n[(CRONOS)]TJ\nET\n" +
		"719 128 m\n719 400 l\n766 400 l\n766 128 l\nh\nW n\n" +
		"q\n4.16667 0 0 4.16667 0 0 cm BT\n/R10 10 Tf\n[(4P 1,3 DRIVE GSE)]TJ\nET\n" +
		"719 1439 m\n719 1500 l\n766 1500 l\n766 1439 l\nh\nW n\n0 0 0 rg\n" +
		"q\n4.16667 0 0 4.16667 0 0 cm BT\n/R10 10 Tf\n[(48410,0)]TJ\nET\n" +
		"719 2530 m\n719 2600 l\n766 2600 l\n766 2530 l\nh\nW n\n" +
		"q\n4.16667 0 0 4.16667 0 0 cm BT\n/R10 10 Tf\n[(13,346)]TJ\nET\n" +
		"719 2688 m\n719 2750 l\n766 2750 l\n766 2688 l\nh\nW n\n" +
		"q\n4.16667 0 0 4.16667 0 0 cm BT\n/R10 10 Tf\n[(12,800)]TJ\nET\n" +
		"800 128 m\n800 400 l\n850 400 l\n850 128 l\nh\nW n\n" +
		"q\n4.16667 0 0 4.16667 0 0 cm BT\n/R10 10 Tf\n[(4P 1,3 DRIVE GSE)]TJ\nET\n" +
		"800 2530 m\n800 2600 l\n850 2600 l\n850 2530 l\nh\nW n\n" +
		"q\n4.16667 0 0 4.16667 0 0 cm BT\n/R10 10 Tf\n[(99,999)]TJ\nET\n"

	pdf := "<<" + ">>\nstream\n" + cmapBold + "endstream\n" +
		"<<" + ">>\nstream\n" + cmapItalic + "endstream\n" +
		"<<" + ">>\nstream\n" + content + "endstream\n"
	observed := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	rows, err := parsePDF([]byte(pdf), observed)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1 (later list line must not add a year)", len(rows))
	}
	row := rows[0]
	if row.Currency != "USD" {
		t.Fatalf("currency = %s", row.Currency)
	}
	if got := row.Keys()[0]; got != "FIAT CRONOS 4P 1,3 DRIVE GSE" {
		t.Fatalf("key = %q", got)
	}
	assertCell(t, row, 2026, 48410)
	assertCell(t, row, 2019, 13346)
	assertCell(t, row, 2018, 12800)
}

func TestParseFile_optionalGuide(t *testing.T) {
	path := os.Getenv("CCA_PDF")
	if path == "" {
		t.Skip("set CCA_PDF to check a real guide")
	}
	observed := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	rows, err := ParseFile(path, observed)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]map[int]int64{
		"TOYOTA COROLLA 4P 2,0 XEI CVT 2024":          {2025: 31993, 2024: 30469, 2023: 29555},
		"FIAT CRONOS 4P 1,3 DRIVE GSE":                {2019: 13346, 2018: 12800},
		"VOLKSWAGEN GOLF 5P 1,4 TSI COMFORTLINE 2018": {2020: 24357, 2019: 23140, 2018: 21068, 2017: 19177},
		"TOYOTA COROLLA 4P 2,0 XEI CVT SAFETY 2025":   {2026: 48410, 2025: 33593},
	}
	found := map[string]guide.Row{}
	for _, row := range rows {
		if len(row.Keys()) == 0 {
			continue
		}
		if _, ok := want[row.Keys()[0]]; ok {
			found[row.Keys()[0]] = row
		}
	}
	for key, years := range want {
		row, ok := found[key]
		if !ok {
			t.Errorf("missing %s", key)
			continue
		}
		if row.Currency != "USD" {
			t.Errorf("%s currency = %s", key, row.Currency)
		}
		for year, price := range years {
			assertCell(t, row, year, price)
		}
	}
}

func assertCell(t *testing.T, row guide.Row, year int, price int64) {
	t.Helper()
	for _, c := range row.Cells {
		if c.Year == year {
			if c.Price != price {
				t.Errorf("%s %d = %d, want %d", row.Keys()[0], year, c.Price, price)
			}
			return
		}
	}
	t.Errorf("%s missing year %d", row.Keys()[0], year)
}
