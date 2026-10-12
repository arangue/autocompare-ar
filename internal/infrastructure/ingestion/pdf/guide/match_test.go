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
