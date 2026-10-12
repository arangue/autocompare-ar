package carsbase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/arangue/autocompare-ar/internal/domain"
)

// Sheet is one thecarsbase trim-year whose calendar year falls inside the
// years printed on the title. YearTo is nil when the title says "presente".
type Sheet struct {
	URL                string
	Brand              string
	Model              string
	Title              string
	Year               int
	YearFrom           int
	YearTo             *int
	Generation         string
	Engine             string
	DisplacementCC     *int
	Horsepower         *int
	Transmission       string
	Fuel               string
	Doors              *int
	Seats              *int
	ConsumptionCity    *float64
	ConsumptionHighway *float64
	ConsumptionMixed   *float64
	TorqueNm           *int
	Cylinders          *int
	TopSpeedKmh        *int
	Traction           string
	Features           []domain.VehicleFeature

	gearbox string
	speeds  string
}

// TrimRef is a catalog trim. GenTo and YearTo are 0 when the span is open.
type TrimRef struct {
	ID       int
	Brand    string
	Model    string
	Name     string
	GenFrom  int
	GenTo    int
	YearFrom int
	YearTo   int
}

// Link attaches one sheet year to a catalog trim.
type Link struct {
	TrimID int
	Year   int
	URL    string
}

var titleYears = regexp.MustCompile(`(19\d{2}|20\d{2})\s*[–-]\s*((?:19|20)\d{2}|[Pp]resente)\s*$`)

// TitleSpan reads the sales period at the end of a thecarsbase title.
// open is true when the title ends in "presente".
func TitleSpan(title string) (from int, to int, open bool, ok bool) {
	m := titleYears.FindStringSubmatch(strings.TrimSpace(title))
	if m == nil {
		return 0, 0, false, false
	}
	from, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, 0, false, false
	}
	if strings.EqualFold(m[2], "presente") {
		return from, 0, true, true
	}
	to, err = strconv.Atoi(m[2])
	if err != nil || to < from {
		return 0, 0, false, false
	}
	return from, to, false, true
}

// ParseDetalles reads Mecánica and Carrocería into spec fields, and Confort,
// Seguridad, and Tecnología into features. The leading "Especificaciones
// completas" dump is ignored so those fields are not counted twice.
func ParseDetalles(raw string) (Sheet, error) {
	var sheet Sheet
	sections, err := orderedObject([]byte(raw))
	if err != nil {
		return Sheet{}, err
	}
	for _, section := range sections {
		if sheet.Generation == "" && (section.key == "Generación" || section.key == "Generacion") {
			sheet.Generation = scalar(section.val)
		}
		fields, err := orderedObject(section.val)
		if err != nil {
			continue
		}
		if sheet.Generation == "" {
			sheet.Generation = generationFrom(fields)
		}
		switch section.key {
		case "Mecánica", "Carrocería":
			applySpec(&sheet, fields)
		case "Confort", "Seguridad", "Tecnología":
			for _, field := range fields {
				value := scalar(field.val)
				if value == "" {
					continue
				}
				code := featureCode(field.key)
				if code == "" {
					continue
				}
				sheet.Features = append(sheet.Features, domain.VehicleFeature{
					Code:     code,
					Name:     field.key,
					Category: section.key,
					Value:    value,
				})
			}
		}
	}
	if sheet.Transmission == "" {
		sheet.Transmission = strings.TrimSpace(sheet.gearbox + " " + sheet.speeds)
	}
	return sheet, nil
}

// matchTrim links sheets to a trim only when exactly one title contains every
// trim-name token and that title's year span sits inside the generation.
// ponytail: ambiguous titles stay unmatched; add an explicit alias when a human picks one.
func matchTrim(trim TrimRef, sheets []Sheet) []Link {
	type group struct {
		from    int
		to      *int
		maxYear int
		byYear  map[int]string
	}
	groups := map[string]*group{}
	for _, sheet := range sheets {
		if fold(sheet.Brand) != fold(trim.Brand) || fold(sheet.Model) != fold(trim.Model) {
			continue
		}
		if !tokensIn(fold(trim.Name), fold(sheet.Title)) {
			continue
		}
		g := groups[sheet.Title]
		if g == nil {
			g = &group{from: sheet.YearFrom, to: sheet.YearTo, byYear: map[int]string{}}
			groups[sheet.Title] = g
		}
		if sheet.Year > g.maxYear {
			g.maxYear = sheet.Year
		}
		g.byYear[sheet.Year] = sheet.URL
	}

	var chosen *group
	for _, g := range groups {
		end := g.maxYear
		if g.to != nil {
			end = *g.to
		}
		if !spanInside(trim.GenFrom, trim.GenTo, g.from, end) {
			continue
		}
		if chosen != nil {
			return nil
		}
		chosen = g
	}
	if chosen == nil {
		return nil
	}

	from, to := trim.YearFrom, trim.YearTo
	if from == 0 {
		from = trim.GenFrom
	}
	if to == 0 {
		to = trim.GenTo
	}
	if to == 0 {
		to = 9999
	}
	var links []Link
	for year, url := range chosen.byYear {
		if year < from || year > to {
			continue
		}
		if year < chosen.from || (chosen.to != nil && year > *chosen.to) {
			continue
		}
		links = append(links, Link{TrimID: trim.ID, Year: year, URL: url})
	}
	return links
}

// MatchAll links each catalog trim to the sheets of its brand and model.
func MatchAll(trims []TrimRef, sheets []Sheet) []Link {
	byModel := map[string][]Sheet{}
	for _, sheet := range sheets {
		key := fold(sheet.Brand) + "\x00" + fold(sheet.Model)
		byModel[key] = append(byModel[key], sheet)
	}
	var links []Link
	for _, trim := range trims {
		key := fold(trim.Brand) + "\x00" + fold(trim.Model)
		links = append(links, matchTrim(trim, byModel[key])...)
	}
	return links
}

func spanInside(genFrom, genTo, titleFrom, titleTo int) bool {
	if genFrom == 0 || titleFrom == 0 || titleTo == 0 {
		return false
	}
	if genTo == 0 {
		genTo = 9999
	}
	return titleFrom >= genFrom && titleTo <= genTo
}

func tokensIn(name, title string) bool {
	nameTokens := strings.Fields(name)
	if len(nameTokens) == 0 {
		return false
	}
	have := map[string]bool{}
	for _, token := range strings.Fields(title) {
		have[token] = true
	}
	for _, token := range nameTokens {
		if !have[token] {
			return false
		}
	}
	return true
}

type field struct {
	key string
	val json.RawMessage
}

func orderedObject(raw []byte) ([]field, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, fmt.Errorf("expected object")
	}
	var fields []field
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, fmt.Errorf("expected string key")
		}
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, err
		}
		fields = append(fields, field{key: key, val: val})
	}
	return fields, nil
}

func scalar(raw json.RawMessage) string {
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return strings.TrimSpace(text)
	}
	return strings.TrimSpace(string(raw))
}

func generationFrom(fields []field) string {
	for _, field := range fields {
		if field.key == "Generación" || field.key == "Generacion" {
			return scalar(field.val)
		}
	}
	return ""
}

func applySpec(sheet *Sheet, fields []field) {
	for _, field := range fields {
		value := scalar(field.val)
		if value == "" {
			continue
		}
		switch field.key {
		case "Motor":
			sheet.Engine = value
		case "Cilindrada (cm³)", "Cilindrada (cm3)":
			if n, ok := parseIntish(value); ok {
				sheet.DisplacementCC = &n
			}
		case "Potencia (hp)":
			if n, ok := parseIntish(value); ok {
				sheet.Horsepower = &n
			}
		case "Tipo de caja":
			sheet.gearbox = value
		case "Velocidades":
			sheet.speeds = value
		case "Combustible":
			sheet.Fuel = value
		case "Consumo en ciudad (l/100 km)":
			if n, ok := parseFloat(value); ok {
				sheet.ConsumptionCity = &n
			}
		case "Consumo en ruta (l/100 km)":
			if n, ok := parseFloat(value); ok {
				sheet.ConsumptionHighway = &n
			}
		case "Consumo mixto (l/100 km)":
			if n, ok := parseFloat(value); ok {
				sheet.ConsumptionMixed = &n
			}
		case "Torque (Nm)":
			if n, ok := parseIntish(value); ok {
				sheet.TorqueNm = &n
			}
		case "Cilindros":
			if n, ok := parseIntish(value); ok {
				sheet.Cylinders = &n
			}
		case "Velocidad máxima (km/h)", "Velocidad maxima (km/h)":
			if n, ok := parseIntish(value); ok {
				sheet.TopSpeedKmh = &n
			}
		case "Tracción", "Traccion":
			sheet.Traction = value
		case "Puertas":
			if n, ok := parseIntish(value); ok {
				sheet.Doors = &n
			}
		case "Plazas":
			if n, ok := parseIntish(value); ok {
				sheet.Seats = &n
			}
		}
	}
}

func parseIntish(s string) (int, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(strings.ToLower(s), "hp")
	s = strings.ReplaceAll(s, " ", "")
	if strings.Count(s, ".") == 1 {
		parts := strings.SplitN(s, ".", 2)
		if len(parts[1]) == 3 && !strings.Contains(s, ",") {
			s = parts[0] + parts[1]
		}
	}
	s = strings.ReplaceAll(s, ",", "")
	n, err := strconv.Atoi(s)
	return n, err == nil && n > 0
}

func parseFloat(s string) (float64, bool) {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", "."))
	n, err := strconv.ParseFloat(s, 64)
	return n, err == nil && n > 0
}

func featureCode(name string) string {
	return strings.ReplaceAll(fold(name), " ", "_")
}

func fold(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	space := true
	for _, r := range s {
		switch r {
		case 'á', 'à', 'ä':
			r = 'a'
		case 'é', 'è', 'ë':
			r = 'e'
		case 'í', 'ì', 'ï':
			r = 'i'
		case 'ó', 'ò', 'ö':
			r = 'o'
		case 'ú', 'ù', 'ü':
			r = 'u'
		case 'ñ':
			r = 'n'
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' {
			b.WriteRune(r)
			space = false
			continue
		}
		if !space {
			b.WriteByte(' ')
			space = true
		}
	}
	return strings.TrimSpace(b.String())
}
