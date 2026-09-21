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
