package carsbase

import (
	"fmt"
	"strings"
	"unicode"
)

// CatalogTrim is one trim taken from a thecarsbase title.
// Years maps a calendar year to the sheet URL kept for that year.
type CatalogTrim struct {
	Brand      string
	Model      string
	Generation string
	Name       string
	YearFrom   int
	YearTo     *int
	Years      map[int]string
}

// BuildCatalog turns sheets into trims. One title is one trim.
// ponytail: thecarsbase has a body generation on almost no sheet, so the
// generation name is that text when present and the model name otherwise.
// Two titles that strip to the same name keep the sales span in the name.
func BuildCatalog(sheets []Sheet) []CatalogTrim {
	type key struct {
		brand, model, gen, base string
		from, to                int
		open                    bool
	}
	type group struct {
		brand, model, gen, base string
		from                    int
		to                      *int
		years                   map[int]string
	}

	groups := map[key]*group{}
	var order []key
	for _, sheet := range sheets {
		base := trimBase(sheet.Title, sheet.Brand, sheet.Model)
		gen := strings.TrimSpace(sheet.Generation)
		if gen == "" {
			gen = sheet.Model
		}
		open := sheet.YearTo == nil
		to := 0
		if !open {
			to = *sheet.YearTo
		}
		k := key{fold(sheet.Brand), fold(sheet.Model), fold(gen), fold(base), sheet.YearFrom, to, open}
		g := groups[k]
		if g == nil {
			var toPtr *int
			if !open {
				v := to
				toPtr = &v
			}
			g = &group{
				brand: sheet.Brand,
				model: sheet.Model,
				gen:   gen,
				base:  base,
				from:  sheet.YearFrom,
				to:    toPtr,
				years: map[int]string{},
			}
			groups[k] = g
			order = append(order, k)
		}
		if _, ok := g.years[sheet.Year]; !ok {
			g.years[sheet.Year] = sheet.URL
		}
	}

	counts := map[string]int{}
	for k := range groups {
		counts[k.brand+"\x00"+k.model+"\x00"+k.gen+"\x00"+k.base]++
	}

	out := make([]CatalogTrim, 0, len(order))
	for _, k := range order {
		g := groups[k]
		name := g.base
		if name == "" {
			name = g.model
		}
		if counts[k.brand+"\x00"+k.model+"\x00"+k.gen+"\x00"+k.base] > 1 {
			name = strings.TrimSpace(name + " " + spanLabel(g.from, g.to))
		}
		out = append(out, CatalogTrim{
			Brand:      g.brand,
			Model:      g.model,
			Generation: g.gen,
			Name:       name,
			YearFrom:   g.from,
			YearTo:     g.to,
			Years:      g.years,
		})
	}
	return out
}

func spanLabel(from int, to *int) string {
	if to == nil {
		return fmt.Sprintf("%d–presente", from)
	}
	return fmt.Sprintf("%d–%d", from, *to)
}

func trimBase(title, brand, model string) string {
	rest := stripPrefixFold(title, brand)
	rest = stripPrefixFold(rest, model)
	rest = strings.TrimSpace(titleYears.ReplaceAllString(rest, ""))
	if rest == "" {
		rest = strings.TrimSpace(titleYears.ReplaceAllString(title, ""))
	}
	return rest
}

func stripPrefixFold(s, prefix string) string {
	pf := fold(prefix)
	if pf == "" {
		return strings.TrimSpace(s)
	}
	for i, r := range s {
		if i == 0 {
			continue
		}
		head := fold(s[:i])
		if head != pf {
			if len(head) > len(pf) {
				break
			}
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		return strings.TrimSpace(s[i:])
	}
	if fold(s) == pf {
		return ""
	}
	return strings.TrimSpace(s)
}
