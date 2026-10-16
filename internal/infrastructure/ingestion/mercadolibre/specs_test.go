package mercadolibre

import "testing"

func TestParseSpecs(t *testing.T) {
	html := []byte(`{"id":"Versión","text":"1.6 Titanium 120cv"},{"id":"Puertas","text":"5"},{"id":"Transmisión","text":"Manual"},{"id":"Tipo de carrocería","text":"Hatchback"}`)
	got := ParseSpecs(html)
	if got.Doors != "5P" || got.Transmission != "Manual" || got.Body != "Hatchback" {
		t.Fatalf("got %+v", got)
	}
	if tok := got.Tokens(); tok != "5P Manual" {
		t.Fatalf("tokens = %q", tok)
	}
}

func TestParseSpecs_escaped(t *testing.T) {
	html := []byte(`{"id":"Puertas","text":"4"},{"id":"Transmisi\u00f3n","text":"Automática"},{"id":"Tipo de carrocer\u00eda","text":"Sedán"}`)
	got := ParseSpecs(html)
	if got.Doors != "4P" || got.Transmission != "Automatica" || got.Body != "Sedan" {
		t.Fatalf("got %+v", got)
	}
}

func TestParseSpecs_bodyWhenNoDoors(t *testing.T) {
	html := []byte(`{"id":"Tipo de carrocería","text":"Hatchback"}`)
	got := ParseSpecs(html)
	if got.Tokens() != "Hatchback" {
		t.Fatalf("tokens = %q", got.Tokens())
	}
}

func TestParseSpecs_empty(t *testing.T) {
	got := ParseSpecs([]byte("<html>no specs</html>"))
	if got.Tokens() != "" {
		t.Fatalf("tokens = %q", got.Tokens())
	}
}

func TestParseSpecs_focusSePlus(t *testing.T) {
	html := []byte(`{"id":"Versión","text":"2.0 Se Plus At6"},{"id":"Puertas","text":"5"},{"id":"Transmisión","text":"Automática"},{"id":"Tipo de carrocería","text":"Hatchback"}`)
	got := ParseSpecs(html)
	if got.Doors != "5P" || got.Transmission != "Automatica" {
		t.Fatalf("got %+v", got)
	}
}

func TestParseSpecs_htmlFallback(t *testing.T) {
	html := []byte(`<p><span>Puertas:</span></span> <span class="ui-pdp-family--SEMIBOLD"><span>5</span></span></p>
<p><span>Transmisión:</span></span> <span class="ui-pdp-family--SEMIBOLD"><span>Manual</span></span></p>
<p><span>Tipo de carrocería:</span></span> <span class="ui-pdp-family--SEMIBOLD"><span>Hatchback</span></span></p>`)
	got := ParseSpecs(html)
	if got.Doors != "5P" || got.Transmission != "Manual" || got.Body != "Hatchback" {
		t.Fatalf("got %+v", got)
	}
	if tok := got.Tokens(); tok != "5P Manual" {
		t.Fatalf("tokens = %q", tok)
	}
}

func TestParseSpecs_htmlAutomatica(t *testing.T) {
	html := []byte(`<span>Puertas:</span></span> <span><span>4</span></span>
<span>Transmisión:</span></span> <span><span>Automática</span></span>`)
	got := ParseSpecs(html)
	if got.Doors != "4P" || got.Transmission != "Automatica" {
		t.Fatalf("got %+v", got)
	}
}
