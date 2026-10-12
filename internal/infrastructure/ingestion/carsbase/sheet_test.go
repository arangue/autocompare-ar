package carsbase

import "testing"

func TestTitleSpan(t *testing.T) {
	from, to, open, ok := TitleSpan("Toyota Corolla XEI 2.0 CVT 4P 2020–presente")
	if !ok || from != 2020 || !open || to != 0 {
		t.Fatalf("presente = %d %d open=%v ok=%v", from, to, open, ok)
	}
	from, to, open, ok = TitleSpan("Toyota Corolla XLI 1.8 CVT 4P 2014–2019")
	if !ok || open || from != 2014 || to != 2019 {
		t.Fatalf("closed = %d %d open=%v ok=%v", from, to, open, ok)
	}
	if _, _, _, ok = TitleSpan("sin años"); ok {
		t.Fatal("expected no span")
	}
}

func TestParseDetallesIgnoresDump(t *testing.T) {
	raw := `{
		"Especificaciones completas del X": {"Motor": "nope", "Airbags (total)": "9"},
		"Mecánica": {
			"Motor": "1.4L Turbo",
			"Potencia (hp)": "165",
			"Tipo de caja": "MT",
			"Velocidades": "5",
			"Combustible": "Nafta",
			"Cilindrada (cm³)": "1368",
			"Consumo en ciudad (l/100 km)": "9.2",
			"Consumo en ruta (l/100 km)": "6.8",
			"Consumo mixto (l/100 km)": "7.5",
			"Torque (Nm)": "200",
			"Cilindros": "4",
			"Tracción": "Delantera",
			"Velocidad máxima (km/h)": "190",
			"Generación": "Mk7"
		},
		"Carrocería": {"Puertas": "3", "Plazas": "4"},
		"Seguridad": {"Airbags (total)": "2", "Frenos ABS": "Sí"},
		"Confort": {"Aire acondicionado": "Sí"}
	}`
	sheet, err := ParseDetalles(raw)
	if err != nil {
		t.Fatal(err)
	}
	if sheet.Engine != "1.4L Turbo" || sheet.Transmission != "MT 5" || sheet.Fuel != "Nafta" || sheet.Generation != "Mk7" {
		t.Fatalf("spec = %+v", sheet)
	}
	if sheet.Horsepower == nil || *sheet.Horsepower != 165 || sheet.DisplacementCC == nil || *sheet.DisplacementCC != 1368 {
		t.Fatalf("numbers = hp %v cc %v", sheet.Horsepower, sheet.DisplacementCC)
	}
	if sheet.Doors == nil || *sheet.Doors != 3 || sheet.Seats == nil || *sheet.Seats != 4 {
		t.Fatalf("body = doors %v seats %v", sheet.Doors, sheet.Seats)
	}
	if sheet.TorqueNm == nil || *sheet.TorqueNm != 200 || sheet.Cylinders == nil || *sheet.Cylinders != 4 || sheet.TopSpeedKmh == nil || *sheet.TopSpeedKmh != 190 || sheet.Traction != "Delantera" {
		t.Fatalf("mechanics = %+v", sheet)
	}
	if sheet.ConsumptionMixed == nil || *sheet.ConsumptionMixed != 7.5 {
		t.Fatalf("mixed = %v", sheet.ConsumptionMixed)
	}
	if len(sheet.Features) != 3 {
		t.Fatalf("features = %+v", sheet.Features)
	}
	if sheet.Features[0].Code != "airbags_total" || sheet.Features[0].Category != "Seguridad" || sheet.Features[0].Value != "2" {
		t.Fatalf("first feature = %+v", sheet.Features[0])
	}
}

func TestMatchAllUniqueTitleInsideGeneration(t *testing.T) {
	to2019 := 2019
	to2026 := 2026
	sheets := []Sheet{
		{URL: "cvt-2020", Brand: "Toyota", Model: "Corolla", Title: "Toyota Corolla XEI 2.0 CVT 4P 2020–presente", Year: 2020, YearFrom: 2020, YearTo: &to2026},
		{URL: "cvt-2022", Brand: "Toyota", Model: "Corolla", Title: "Toyota Corolla XEI 2.0 CVT 4P 2020–presente", Year: 2022, YearFrom: 2020, YearTo: &to2026},
		{URL: "mt-2020", Brand: "Toyota", Model: "Corolla", Title: "Toyota Corolla XEI 2.0 MT 4P 2020–2025", Year: 2020, YearFrom: 2020, YearTo: &to2026},
		{URL: "old-2019", Brand: "Toyota", Model: "Corolla", Title: "Toyota Corolla XLI 1.8 CVT 4P 2014–2019", Year: 2019, YearFrom: 2014, YearTo: &to2019},
		{URL: "dsg", Brand: "Volkswagen", Model: "Golf", Title: "Volkswagen Golf Comfortline 1.4 TSI DSG 5P 2018–2019", Year: 2018, YearFrom: 2018, YearTo: &to2019},
		{URL: "mt", Brand: "Volkswagen", Model: "Golf", Title: "Volkswagen Golf Comfortline 1.4 TSI MT 5P 2018–2019", Year: 2018, YearFrom: 2018, YearTo: &to2019},
	}
	trims := []TrimRef{
		{ID: 1, Brand: "Toyota", Model: "Corolla", Name: "XEi 2.0 CVT", GenFrom: 2019, GenTo: 2026, YearFrom: 2019, YearTo: 2026},
		{ID: 2, Brand: "Toyota", Model: "Corolla", Name: "XLi 1.8 CVT", GenFrom: 2019, GenTo: 2026, YearFrom: 2019, YearTo: 2026},
		{ID: 3, Brand: "Volkswagen", Model: "Golf", Name: "Comfortline", GenFrom: 2014, GenTo: 2020, YearFrom: 2014, YearTo: 2020},
	}
	links := MatchAll(trims, sheets)
	if len(links) != 2 {
		t.Fatalf("links = %+v", links)
	}
	for _, link := range links {
		if link.TrimID != 1 || (link.Year != 2020 && link.Year != 2022) {
			t.Fatalf("unexpected link %+v", link)
		}
	}
}
