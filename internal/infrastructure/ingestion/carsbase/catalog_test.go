package carsbase

import "testing"

func TestBuildCatalogFromTitles(t *testing.T) {
	to2008 := 2008
	to2013 := 2013
	sheets := []Sheet{
		{URL: "a2000", Brand: "Ford", Model: "Focus", Title: "Ford Focus Ambiente 1.6 MT 4P 2000–2008", Year: 2000, YearFrom: 2000, YearTo: &to2008},
		{URL: "a2001", Brand: "Ford", Model: "Focus", Title: "Ford Focus Ambiente 1.6 MT 4P 2000–2008", Year: 2001, YearFrom: 2000, YearTo: &to2008},
		{URL: "g2008", Brand: "Ford", Model: "Focus", Title: "Ford Focus Ghia 2.0 MT 4P 2008–2013", Year: 2008, YearFrom: 2008, YearTo: &to2013},
		{URL: "g2008b", Brand: "Ford", Model: "Focus", Title: "Ford Focus Ghia 2.0 MT 4P 2008–2013", Year: 2008, YearFrom: 2008, YearTo: &to2013},
	}
	got := BuildCatalog(sheets)
	if len(got) != 2 {
		t.Fatalf("trims = %+v", got)
	}
	if got[0].Brand != "Ford" || got[0].Model != "Focus" || got[0].Generation != "Focus" || got[0].Name != "Ambiente 1.6 MT 4P" {
		t.Fatalf("ambiente = %+v", got[0])
	}
	if got[0].YearFrom != 2000 || got[0].YearTo == nil || *got[0].YearTo != 2008 || len(got[0].Years) != 2 {
		t.Fatalf("ambiente years = %+v", got[0])
	}
	if got[1].Name != "Ghia 2.0 MT 4P" || got[1].Years[2008] != "g2008" {
		t.Fatalf("ghia = %+v", got[1])
	}
}

func TestBuildCatalogKeepsSameNameWithDifferentSpans(t *testing.T) {
	to2019 := 2019
	sheets := []Sheet{
		{URL: "old", Brand: "Toyota", Model: "Corolla", Title: "Toyota Corolla XEI 2.0 CVT 4P 2014–2019", Year: 2014, YearFrom: 2014, YearTo: &to2019},
		{URL: "new", Brand: "Toyota", Model: "Corolla", Title: "Toyota Corolla XEI 2.0 CVT 4P 2020–presente", Year: 2020, YearFrom: 2020, Generation: "E210"},
	}
	got := BuildCatalog(sheets)
	if len(got) != 2 {
		t.Fatalf("trims = %+v", got)
	}
	if got[0].Generation != "Corolla" || got[0].Name != "XEI 2.0 CVT 4P" || got[0].YearTo == nil || *got[0].YearTo != 2019 {
		t.Fatalf("old = %+v", got[0])
	}
	if got[1].Generation != "E210" || got[1].Name != "XEI 2.0 CVT 4P" || got[1].YearTo != nil {
		t.Fatalf("new = %+v", got[1])
	}
}

func TestBuildCatalogDisambiguatesSameGeneration(t *testing.T) {
	to2019 := 2019
	sheets := []Sheet{
		{URL: "old", Brand: "Toyota", Model: "Corolla", Title: "Toyota Corolla XEI 2.0 CVT 4P 2014–2019", Year: 2014, YearFrom: 2014, YearTo: &to2019},
		{URL: "new", Brand: "Toyota", Model: "Corolla", Title: "Toyota Corolla XEI 2.0 CVT 4P 2020–presente", Year: 2020, YearFrom: 2020},
	}
	got := BuildCatalog(sheets)
	if len(got) != 2 || got[0].Name != "XEI 2.0 CVT 4P 2014–2019" || got[1].Name != "XEI 2.0 CVT 4P 2020–presente" {
		t.Fatalf("%+v", got)
	}
}

func TestTrimBaseDoesNotStripInsideAToken(t *testing.T) {
	if got := trimBase("Ford GTI 2.0 MT 2000–2008", "Ford", "GT"); got != "GTI 2.0 MT" {
		t.Fatalf("got %q", got)
	}
}
