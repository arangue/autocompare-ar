package guide

import "testing"

func TestMatch_guideLine(t *testing.T) {
	y := func(years ...int) map[int]struct{} {
		m := map[int]struct{}{}
		for _, year := range years {
			m[year] = struct{}{}
		}
		return m
	}
	catalog := []Trim{
		{ID: 1622, Brand: "Ford", Model: "Focus", Name: "S 1.6 MT 5P", Years: y(2019)},
		{ID: 1621, Brand: "Ford", Model: "Focus", Name: "S 1.6 MT 4P", Years: y(2019)},
		{ID: 1700, Brand: "Ford", Model: "Focus IV", Name: "S 1.6 MT 5P", Years: y(2019)},
		{ID: 4351, Brand: "Toyota", Model: "Corolla", Name: "XEI 1.8 MT 4P", Years: y(2012, 2015)},
		{ID: 4355, Brand: "Toyota", Model: "Corolla", Name: "XEI Pack 1.8 MT 4P", Years: y(2015)},
		{ID: 4356, Brand: "Toyota", Model: "Corolla", Name: "XEi 1.8 AT 4P", Years: y(2012)},
		{ID: 4360, Brand: "Toyota", Model: "Corolla", Name: "XEI 2.0 CVT 4P 2014–2019", Years: y(2016)},
		{ID: 4361, Brand: "Toyota", Model: "Corolla", Name: "XEI 2.0 CVT 4P 2020–presente", Years: y(2021)},
	}

	cases := []struct {
		brand, model, trim string
		year               int
		id                 int
		ok                 bool
	}{
		{"FORD", "Focus III", "1.6 S MT (125cv) 5Ptas. (L15)", 2019, 1622, true},
		{"FORD", "Focus III", "1.6 S MT (125cv) 4Ptas. (L15)", 2019, 1621, true},
		{"FORD", "Focus III", "1.6 S MT", 2019, 0, false},
		{"FORD", "Focus IV", "1.6 S MT 5Ptas.", 2019, 1700, true},
		{"TOYOTA", "XEI", "1.8 MT (136cv) (L12)", 2012, 4351, true},
		{"TOYOTA", "Corolla", "XEI 1.8 MT", 2012, 4351, true},
		{"TOYOTA", "Corolla", "XEI Pack 1.8 MT", 2015, 4355, true},
		{"TOYOTA", "XEI", "1.8 MT", 2011, 0, false},
		{"TOYOTA", "COROLLA", "4P 1,8 XEI", 2012, 0, false},
		{"FORD", "HYUNDAI", "Accent 1.6", 2019, 0, false},
		{"TOYOTA", "Corolla", "XEI 2.0 CVT", 2016, 4360, true},
		{"TOYOTA", "Corolla", "XEI 2.0 CVT", 2021, 4361, true},
	}
	for _, tc := range cases {
		id, ok := Match(tc.brand, tc.model, tc.trim, tc.year, catalog)
		if id != tc.id || ok != tc.ok {
			t.Errorf("%s %s %s %d = %d %v, want %d %v", tc.brand, tc.model, tc.trim, tc.year, id, ok, tc.id, tc.ok)
		}
	}
}

func TestMatchTitle(t *testing.T) {
	y := func(years ...int) map[int]struct{} {
		m := map[int]struct{}{}
		for _, year := range years {
			m[year] = struct{}{}
		}
		return m
	}
	catalog := []Trim{
		{ID: 1622, Brand: "Ford", Model: "Focus", Name: "S 1.6 MT 5P", Years: y(2019)},
		{ID: 1621, Brand: "Ford", Model: "Focus", Name: "S 1.6 MT 4P", Years: y(2019)},
		{ID: 1700, Brand: "Ford", Model: "Focus IV", Name: "S 1.6 MT 5P", Years: y(2019)},
		{ID: 4351, Brand: "Toyota", Model: "Corolla", Name: "XEI 1.8 MT 4P", Years: y(2012, 2015)},
		{ID: 4355, Brand: "Toyota", Model: "Corolla", Name: "XEI Pack 1.8 MT 4P", Years: y(2015)},
		{ID: 4356, Brand: "Toyota", Model: "Corolla", Name: "XEi 1.8 AT 4P", Years: y(2012)},
		{ID: 4360, Brand: "Toyota", Model: "Corolla", Name: "XEI 2.0 CVT 4P 2014–2019", Years: y(2016)},
		{ID: 4361, Brand: "Toyota", Model: "Corolla", Name: "XEI 2.0 CVT 4P 2020–presente", Years: y(2021)},
		{ID: 1800, Brand: "Ford", Model: "Focus", Name: "SE Plus 2.0 AT 5P", Years: y(2016)},
		{ID: 1801, Brand: "Ford", Model: "Focus", Name: "SE Plus 2.0 AT 4P", Years: y(2016)},
		{ID: 1900, Brand: "Ford", Model: "EcoSport", Name: "Freestyle 1.6 MT 5P", Years: y(2016)},
		{ID: 1901, Brand: "Ford", Model: "EcoSport", Name: "Freestyle 2.0 4x4 MT 5P", Years: y(2016)},
		{ID: 2000, Brand: "Ford", Model: "Territory", Name: "Titanium 1.8T AT 5P", Years: y(2026)},
		{ID: 2001, Brand: "Ford", Model: "Territory", Name: "SEL 1.8T AT 5P", Years: y(2026)},
		{ID: 2100, Brand: "Ford", Model: "Ranger", Name: "Limited 3.2D AT 4P", Years: y(2022)},
		{ID: 2101, Brand: "Ford", Model: "Ranger", Name: "XLT 3.2D AT 4P", Years: y(2022)},
		{ID: 2102, Brand: "Ford", Model: "Ranger", Name: "XLT 3.2D MT 4P", Years: y(2017)},
		{ID: 2103, Brand: "Ford", Model: "Ranger", Name: "XLT 3.2D AT 4P", Years: y(2017)},
		{ID: 2200, Brand: "Ford", Model: "Bronco Sport", Name: "Badlands 2.0T AT 5P", Years: y(2026)},
		{ID: 2201, Brand: "Ford", Model: "Bronco Sport", Name: "Big Bend 1.5T AT 5P", Years: y(2026)},
		{ID: 3000, Brand: "Ford", Model: "Ka", Name: "S 1.5 MT 5P", Years: y(2017)},
		{ID: 3001, Brand: "Ford", Model: "Ka", Name: "SE 1.5 MT 5P", Years: y(2017)},
		{ID: 3002, Brand: "Ford", Model: "Ka", Name: "SEL 1.5 MT 5P", Years: y(2017)},
		{ID: 3010, Brand: "Ford", Model: "Ka", Name: "SEL 1.5 AT 5P", Years: y(2019)},
		{ID: 3011, Brand: "Ford", Model: "Ka", Name: "Freestyle SEL 1.5 AT 5P", Years: y(2019)},
		{ID: 3012, Brand: "Ford", Model: "Ka Sedan", Name: "SEL 1.5 AT 4P", Years: y(2019)},
		{ID: 4001, Brand: "Ford", Model: "Transit", Name: "Van L2H3 2.0D MT", Years: y(2026)},
		{ID: 4002, Brand: "Ford", Model: "Transit", Name: "Van L2H3 2.0D AT", Years: y(2026)},
		{ID: 4003, Brand: "Ford", Model: "Transit", Name: "Van L2H2 2.0D MT", Years: y(2026)},
		{ID: 4102, Brand: "Ford", Model: "Territory", Name: "Titanium 1.5T AT 5P", Years: y(2021)},
		{ID: 4103, Brand: "Ford", Model: "Territory", Name: "Trend 1.5T AT 5P", Years: y(2021)},
	}

	cases := []struct {
		brand, title string
		year         int
		id           int
		ok           bool
	}{
		{"Ford", "Ford Focus S 1.6 MT 5P", 2019, 1622, true},
		{"Ford", "Ford Focus III S 1.6 MT 5P", 2019, 1622, true},
		{"Ford", "Ford Focus 2019 S 1.6 MT 5P", 2019, 1622, true},
		{"Ford", "Ford Focus S 1.6 MT 5P Hibrida", 2019, 0, false},
		{"Ford", "Ford Focus S 1.6 MT", 2019, 0, false},
		{"Ford", "Ford Focus S 1.6 MT 5P", 2018, 0, false},
		{"Ford", "Ford Focus Iii 2016 2.0 Se Plus At6 5P", 2016, 1800, true},
		{"Ford", "Ford Focus Iii 2016 2.0 Se Plus At6", 2016, 0, false},
		{"Ford", "Ford EcoSport 2016 1.6 Freestyle 110cv 4x2", 2016, 1900, true},
		{"Ford", "Ford Territory 2026 1.8 Ecoboost Titanium At", 2026, 2000, true},
		{"Ford", "Ford Ranger 2022 3.2 Cd Limited Tdci 200cv Automatica", 2022, 2100, true},
		{"Ford", "Ford Ranger 2017 3.2 Cd Xlt Tdci 200cv Manual 4x4", 2017, 2102, true},
		{"Ford", "Ford Bronco Sport 2026 2.0 Ecoboost 4wd Badlands At8", 2026, 2200, true},
		{"Ford", "Ford Ka 2017 1.5 Se", 2017, 3001, true},
		{"Ford", "Ford Ka 2019 1.5 Sel At 5p", 2019, 3010, true},
		{"Ford", "Ford Ka 2019 1.5 Sel At 4p", 2019, 3012, true},
		{"Ford", "Ford Transit 2026 2.0 Furgon Mediano Te (l2h3) Manual", 2026, 4001, true},
		{"Ford", "Ford Transit 2026 2.0 Furgon Mediano Tb (l2h2) Manual", 2026, 4003, true},
		{"Ford", "Ford Transit 2011 2.4 Furgon Corto Tdci Aa Tm", 2011, 0, false},
		{"Ford", "Ford Transit 2026 2.0 Furgon Mediano Te (l2h3)", 2026, 0, false},
		{"Ford", "Ford Territory 2026 1.8 Ecoboost Platinum At", 2026, 0, false},
		{"Ford", "Ford Territory 2021 1.5t Sel", 2021, 0, false},
		{"Toyota", "Toyota Corolla XEI 2.0 CVT", 2016, 4360, true},
	}
	for _, tc := range cases {
		id, ok := MatchTitle(tc.brand, tc.title, tc.year, catalog)
		if id != tc.id || ok != tc.ok {
			t.Errorf("%s %q %d = %d %v, want %d %v", tc.brand, tc.title, tc.year, id, ok, tc.id, tc.ok)
		}
	}
}
