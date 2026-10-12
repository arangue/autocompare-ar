package guide

import (
	"regexp"
	"strings"

	"github.com/arangue/autocompare-ar/internal/infrastructure/ingestion/normalize"
)

// Trim is one catalog trim plus the calendar years that have an equipment sheet.
type Trim struct {
	ID    int
	Brand string
	Model string
	Name  string
	Years map[int]struct{}
}

// Match links one guide line to a catalog trim.
// ok is false when zero or several trims fit, or the year has no sheet on that trim.
// ponytail: token subset against one brand. A scorer if this skips too many real lines.
func Match(brand, model, trim string, year int, catalog []Trim) (int, bool) {
	key := normalize.Title(brand)
	best := -1
	for _, item := range catalog {
		if normalize.Title(item.Brand) != key {
			continue
		}
		if _, ok := item.Years[year]; !ok {
			continue
		}
		if !modelOK(model, item.Model) {
			continue
		}
		if n := len(tokens(item.Model)); n > best {
			best = n
		}
	}
	var hits []int
	for _, item := range catalog {
		if normalize.Title(item.Brand) != key {
			continue
		}
		if _, ok := item.Years[year]; !ok {
			continue
		}
		if best >= 0 && (!modelOK(model, item.Model) || len(tokens(item.Model)) != best) {
			continue
		}
		if !fits(model, trim, item) {
			continue
		}
		hits = append(hits, item.ID)
		if len(hits) > 1 {
			return 0, false
		}
	}
	if len(hits) != 1 {
		return 0, false
	}
	return hits[0], true
}

func fits(guideModel, guideTrim string, item Trim) bool {
	want := tokens(guideTrim)
	if modelOK(guideModel, item.Model) {
		want = append(tokens(guideModel)[len(tokens(item.Model)):], want...)
		want = dropRoman(want)
	} else {
		want = append(tokens(guideModel), want...)
	}
	got := tokens(item.Name)
	if len(want) == 0 || len(got) == 0 {
		return false
	}
	wantSet := set(want)
	gotSet := set(got)
	for t := range wantSet {
		if _, ok := gotSet[t]; !ok {
			return false
		}
	}
	guideDoor := false
	for t := range wantSet {
		if isDoor(t) {
			guideDoor = true
			break
		}
	}
	for t := range gotSet {
		if _, ok := wantSet[t]; ok {
			continue
		}
		if isDoor(t) && !guideDoor {
			continue
		}
		return false
	}
	return true
}

func modelOK(guideModel, catModel string) bool {
	g, c := tokens(guideModel), tokens(catModel)
	if len(c) == 0 || len(g) < len(c) {
		return false
	}
	for i := range c {
		if g[i] != c[i] {
			return false
		}
	}
	for _, t := range g[len(c):] {
		if !isRoman(t) {
			return false
		}
	}
	return true
}

func tokens(s string) []string {
	s = stripParens(s)
	var out []string
	for _, f := range strings.Fields(s) {
		f = strings.Trim(f, ".,;:")
		f = strings.ReplaceAll(f, ",", ".")
		f = normalize.Title(f)
		switch f {
		case "2PTAS", "2PTA":
			f = "2P"
		case "3PTAS", "3PTA":
			f = "3P"
		case "4PTAS", "4PTA", "4PUERTAS":
			f = "4P"
		case "5PTAS", "5PTA", "5PUERTAS":
			f = "5P"
		}
		if f == "" || f == "-" || yearSpan.MatchString(f) {
			continue
		}
		out = append(out, f)
	}
	return out
}

var yearSpan = regexp.MustCompile(`^(?:19|20)\d{2}[–-](?:(?:19|20)\d{2}|PRESENTE)$`)

func dropRoman(in []string) []string {
	out := in[:0]
	for _, t := range in {
		if isRoman(t) {
			continue
		}
		out = append(out, t)
	}
	return out
}

func isRoman(t string) bool {
	switch t {
	case "I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X":
		return true
	default:
		return false
	}
}

func isDoor(t string) bool {
	switch t {
	case "2P", "3P", "4P", "5P":
		return true
	default:
		return false
	}
}

func set(in []string) map[string]struct{} {
	out := make(map[string]struct{}, len(in))
	for _, t := range in {
		out[t] = struct{}{}
	}
	return out
}
