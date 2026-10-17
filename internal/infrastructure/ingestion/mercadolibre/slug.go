package mercadolibre

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Slug is a Mercado Libre autos URL segment: lowercase, no accents, hyphenated.
func Slug(s string) string {
	s = norm.NFD.String(strings.TrimSpace(s))
	var b strings.Builder
	hyphen := false
	for _, r := range s {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		r = unicode.ToLower(r)
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			hyphen = false
			continue
		}
		if !hyphen && b.Len() > 0 {
			b.WriteByte('-')
			hyphen = true
		}
	}
	return strings.Trim(b.String(), "-")
}

// SearchPath is brand-model on autos.mercadolibre.com.ar, for example fiat-siena.
func SearchPath(brand, model string) string {
	a, b := Slug(brand), Slug(model)
	switch {
	case a == "":
		return b
	case b == "":
		return a
	default:
		return a + "-" + b
	}
}
