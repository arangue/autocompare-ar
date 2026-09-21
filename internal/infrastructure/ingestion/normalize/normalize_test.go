package normalize

import "testing"

func TestTitle(t *testing.T) {
	cases := []struct{ in, want string }{
		{"toyota  corolla   xei 2.0 cvt", "TOYOTA COROLLA XEI 2.0 CVT"},
		{"Toyota Corolla XEI 2.0 CVT", "TOYOTA COROLLA XEI 2.0 CVT"},
		{"  Corolla XEi  ", "COROLLA XEI"},
		{"Año", "ANO"},
		{"", ""},
		{"  corolla\txei  ", "COROLLA XEI"},
		{"Ñandú", "NANDU"},
		{"años", "ANOS"},
		{"Camión", "CAMION"},
		{"pingüino", "PINGUINO"},
	}
	for _, c := range cases {
		if got := Title(c.in); got != c.want {
			t.Fatalf("Title(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

// Seed aliases from migrations/006. No DB, no network.
var xeiAliases = map[string]string{
	"TOYOTA COROLLA XEI 2.0 CVT": "XEi 2.0 CVT",
	"TOYOTA COROLLA XEI 2.0":     "XEi 2.0 CVT",
	"COROLLA XEI 2.0 CVT":        "XEi 2.0 CVT",
}

func TestTitle_realListingVariants(t *testing.T) {
	cases := []struct {
		in, wantNorm, wantTrim string
	}{
		{"Toyota Corolla XEI 2.0 CVT", "TOYOTA COROLLA XEI 2.0 CVT", "XEi 2.0 CVT"},
		{"toyota  corolla   xei 2.0 cvt", "TOYOTA COROLLA XEI 2.0 CVT", "XEi 2.0 CVT"},
		{"\tToyota Corolla XEi 2.0 CVT  ", "TOYOTA COROLLA XEI 2.0 CVT", "XEi 2.0 CVT"},
		{"TOYOTA COROLLA XEI 2.0 CVT", "TOYOTA COROLLA XEI 2.0 CVT", "XEi 2.0 CVT"},
		{"Toyota Corolla XEI 2.0", "TOYOTA COROLLA XEI 2.0", "XEi 2.0 CVT"},
		{"  toyota corolla xei 2.0", "TOYOTA COROLLA XEI 2.0", "XEi 2.0 CVT"},
		{"Toyota  Corolla  XEI  2.0", "TOYOTA COROLLA XEI 2.0", "XEi 2.0 CVT"},
		{"Corolla XEI 2.0 CVT", "COROLLA XEI 2.0 CVT", "XEi 2.0 CVT"},
		{"corolla   xei 2.0 cvt", "COROLLA XEI 2.0 CVT", "XEi 2.0 CVT"},
		{"  Corolla XEi 2.0 CVT", "COROLLA XEI 2.0 CVT", "XEi 2.0 CVT"},
		{"Toyota Corolla 2.0 XEI CVT 2019", "TOYOTA COROLLA 2.0 XEI CVT 2019", ""},
		{"XEi Pack", "XEI PACK", ""},
		{"VW Golf Comfortline 2019", "VW GOLF COMFORTLINE 2019", ""},
		{"vendo corolla usado", "VENDO COROLLA USADO", ""},
		{"Toyota Hilux SR", "TOYOTA HILUX SR", ""},
	}
	if len(cases) != 15 {
		t.Fatalf("need 15 cases, got %d", len(cases))
	}

	matched := 0
	for _, c := range cases {
		got := Title(c.in)
		if got != c.wantNorm {
			t.Fatalf("Title(%q)=%q want %q", c.in, got, c.wantNorm)
		}
		trim, ok := xeiAliases[got]
		if c.wantTrim == "" {
			if ok {
				t.Fatalf("Title(%q) unexpectedly resolved to %q", c.in, trim)
			}
			continue
		}
		if !ok || trim != c.wantTrim {
			t.Fatalf("Title(%q) resolved %q want %q", c.in, trim, c.wantTrim)
		}
		matched++
	}
	if matched < 10 {
		t.Fatalf("XEi matches = %d, want ≥10", matched)
	}
}
