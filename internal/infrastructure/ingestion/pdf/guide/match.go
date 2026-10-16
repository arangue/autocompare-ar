package guide

import (
	"regexp"
	"strconv"
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

// MatchTitle links one classified title to a catalog trim.
// Mercado Libre titles use AT6, Tdci, EcoBoost; carsbase uses AT, 3.2D, 1.8T.
// ok is false when zero or several trims still fit after that normalization.
func MatchTitle(brand, title string, year int, catalog []Trim) (int, bool) {
	key := normalize.Title(brand)
	rest := dropYear(tokens(title), year)
	rest = appendVanSize(title, rest)
	brandTokens := tokens(brand)
	if len(brandTokens) > 0 && hasPrefix(rest, brandTokens) {
		rest = rest[len(brandTokens):]
	}
	rest = dropRoman(rest)
	titleToks := classifiedTokens(rest)
	if len(titleToks) == 0 {
		return 0, false
	}

	bestScore := -1
	bestExtra := 1 << 20
	var hits []int
	for _, item := range catalog {
		if normalize.Title(item.Brand) != key {
			continue
		}
		if _, ok := item.Years[year]; !ok {
			continue
		}
		modelToks := tokens(item.Model)
		if !modelMatch(rest, titleToks, modelToks) {
			continue
		}
		cat := classifiedTokens(append(append([]string{}, dropOptionalModel(modelToks)...), tokens(item.Name)...))
		if contradict(titleToks, cat) {
			continue
		}
		score := overlap(titleToks, cat)
		if score == 0 {
			continue
		}
		extra := extraVersion(cat, titleToks)
		if score > bestScore || (score == bestScore && extra < bestExtra) {
			bestScore = score
			bestExtra = extra
			hits = hits[:0]
			hits = append(hits, item.ID)
			continue
		}
		if score == bestScore && extra == bestExtra {
			hits = append(hits, item.ID)
		}
	}
	if len(hits) != 1 {
		return 0, false
	}
	return hits[0], true
}

func dropYear(in []string, year int) []string {
	yearToken := strconv.Itoa(year)
	out := make([]string, 0, len(in))
	for _, token := range in {
		if token == yearToken {
			continue
		}
		out = append(out, token)
	}
	return out
}

func dropOptionalModel(in []string) []string {
	out := make([]string, 0, len(in))
	for _, t := range in {
		switch t {
		case "DESIGN", "HATCHBACK":
			continue
		}
		out = append(out, t)
	}
	return out
}

func modelMatch(title, titleNorm, model []string) bool {
	model = dropOptionalModel(model)
	if len(model) == 0 {
		return false
	}
	if hasPrefix(title, model) {
		return true
	}
	if model[len(model)-1] != "SEDAN" {
		return false
	}
	if !hasPrefix(title, model[:len(model)-1]) {
		return false
	}
	return hasToken(title, "SEDAN") || hasToken(titleNorm, "4P")
}

func appendVanSize(title string, toks []string) []string {
	seen := set(toks)
	for _, m := range vanSize.FindAllString(title, -1) {
		t := normalize.Title(m)
		if _, ok := seen[t]; ok {
			continue
		}
		toks = append(toks, t)
		seen[t] = struct{}{}
	}
	return toks
}

func classifiedTokens(in []string) []string {
	diesel := false
	turbo := false
	van := false
	var raw []string
	for _, t := range in {
		t = strings.TrimSuffix(t, "+")
		if strings.HasSuffix(t, "L") && engineNum.MatchString(strings.TrimSuffix(t, "L")) {
			t = strings.TrimSuffix(t, "L")
		}
		if vanSize.MatchString(t) {
			van = true
			diesel = true
		}
		switch t {
		case "AT4", "AT6", "AT8", "AT10", "AUTOMATICA", "AUTOMATICO", "AUTO":
			t = "AT"
		case "MANUAL":
			t = "MT"
		case "TM":
			if van {
				t = "MT"
			}
		case "TA":
			if van {
				t = "AT"
			}
		case "4WD", "AWD":
			t = "4X4"
		case "HIBRIDA", "HIBRIDO", "HYBRID", "HEV", "POWERBOOST":
			t = "HIBRIDA"
		case "ECOBOOST":
			turbo = true
			continue
		case "TDCI", "TDI", "DIESEL":
			diesel = true
			continue
		case "SEDAN":
			t = "4P"
		case "HATCH", "HATCHBACK":
			t = "5P"
		case "FURGON", "FURGÓN", "VAN":
			van = true
			diesel = true
			t = "VAN"
		case "TE", "TB":
			continue
		case "CV", "HP", "CD", "MY25", "MY24", "MY23", "MY22", "LITROS", "MEDIANO", "CORTO", "LARGO":
			continue
		}
		if cvNum.MatchString(t) {
			continue
		}
		if t == "" || t == "-" {
			continue
		}
		raw = append(raw, t)
	}
	out := make([]string, 0, len(raw))
	for _, t := range raw {
		if turbo {
			if e, ok := engineBase(t); ok && !strings.HasSuffix(t, "T") && !strings.HasSuffix(t, "D") {
				t = e + "T"
			}
		}
		if diesel {
			if e, ok := engineBase(t); ok && !strings.HasSuffix(t, "D") {
				t = e + "D"
			}
		}
		out = append(out, t)
	}
	return out
}

func contradict(title, cat []string) bool {
	catSet := set(cat)
	if hasToken(title, "AT") && !hasToken(cat, "AT") && hasToken(cat, "MT") {
		return true
	}
	if hasToken(title, "MT") && !hasToken(cat, "MT") && hasToken(cat, "AT") {
		return true
	}
	if hasToken(title, "4X4") && !hasToken(cat, "4X4") && hasToken(cat, "4X2") {
		return true
	}
	if hasToken(title, "HIBRIDA") && !hasToken(cat, "HIBRIDA") {
		return true
	}
	if hasToken(title, "4P") && hasToken(cat, "5P") && !hasToken(cat, "4P") {
		return true
	}
	if hasToken(title, "5P") && hasToken(cat, "4P") && !hasToken(cat, "5P") {
		return true
	}
	if te, ok := firstEngine(title); ok {
		if ce, ok := firstEngine(cat); ok && te != ce {
			return true
		}
	}
	for _, t := range title {
		if noiseToken(t) {
			continue
		}
		if engineBaseOf(t) != "" {
			continue
		}
		switch t {
		case "AT", "MT", "4X4", "4X2", "4P", "5P", "2P", "3P", "V6", "V8":
			if _, ok := catSet[t]; !ok && !compatMissing(t, catSet) {
				return true
			}
		}
	}
	return false
}

func compatMissing(t string, cat map[string]struct{}) bool {
	switch t {
	case "4P", "5P", "2P", "3P":
		return true
	case "4X4", "4X2":
		_, a := cat["4X4"]
		_, b := cat["4X2"]
		return !a && !b
	}
	return false
}

func extraVersion(cat, title []string) int {
	titleSet := set(title)
	n := 0
	for _, t := range cat {
		if _, ok := titleSet[t]; ok {
			continue
		}
		if engineBaseOf(t) != "" {
			continue
		}
		switch t {
		case "AT", "MT", "4P", "5P", "2P", "3P", "4X4", "4X2", "V6", "V8", "CVT":
			continue
		}
		n++
	}
	return n
}

func noiseToken(t string) bool {
	switch t {
	case "4X2", "LUXURY":
		return true
	}
	return false
}

func overlap(title, cat []string) int {
	catSet := set(cat)
	n := 0
	seen := map[string]bool{}
	for _, t := range title {
		if seen[t] {
			continue
		}
		if _, ok := catSet[t]; ok {
			seen[t] = true
			n++
			continue
		}
		if e, ok := engineBase(t); ok {
			if _, ok := catSet[e]; ok {
				seen[t] = true
				n++
				continue
			}
			if _, ok := catSet[e+"T"]; ok && (t == e || t == e+"T") {
				seen[t] = true
				n++
				continue
			}
			if _, ok := catSet[e+"D"]; ok && (t == e || t == e+"D") {
				seen[t] = true
				n++
			}
		}
	}
	return n
}

func hasToken(in []string, want string) bool {
	for _, t := range in {
		if t == want {
			return true
		}
	}
	return false
}

func firstEngine(in []string) (string, bool) {
	for _, t := range in {
		if e := engineBaseOf(t); e != "" {
			return e, true
		}
	}
	return "", false
}

func engineBase(t string) (string, bool) {
	e := engineBaseOf(t)
	return e, e != ""
}

func engineBaseOf(t string) string {
	t = strings.TrimSuffix(strings.TrimSuffix(t, "T"), "D")
	if engineNum.MatchString(t) {
		return t
	}
	return ""
}

var (
	cvNum     = regexp.MustCompile(`^\d{2,4}CV$`)
	engineNum = regexp.MustCompile(`^\d+\.\d+$`)
	vanSize   = regexp.MustCompile(`(?i)\bL[123]H[123]\b`)
)

func hasPrefix(tokens, prefix []string) bool {
	if len(prefix) > len(tokens) {
		return false
	}
	for i := range prefix {
		if tokens[i] != prefix[i] {
			return false
		}
	}
	return true
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
