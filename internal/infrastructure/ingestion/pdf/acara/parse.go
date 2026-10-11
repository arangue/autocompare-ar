package acara

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/arangue/autocompare-ar/internal/infrastructure/ingestion/pdf/guide"
)

// Row and Cell are the shared guide shape.
type (
	Row  = guide.Row
	Cell = guide.Cell
)

// ParseFile reads an ACARA Guía Oficial de Precios PDF.
func ParseFile(path string, observed time.Time) ([]Row, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	text := ExtractText(data)
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("no text in %s", path)
	}
	return ParseText(text, observed)
}

var (
	headerRe = regexp.MustCompile(`Gu[ií]a Oficial de Precios\s+(.+?)\s+0km\s+((?:20\d\d\s+){5,})`)
	yearRe   = regexp.MustCompile(`20\d\d`)
	priceRe  = regexp.MustCompile(`^(?:-|\d{1,3}(?:\.\d{3})+(?:,\d{2})?|\d+(?:,\d{2})?)`)
)

// ParseText parses the reading-order text of an ACARA guide.
// ponytail: model is the leading words until a displacement or body word (Hybrid, Variant).
// Two guide lines that normalize to the same alias overwrite the same trim/year on upsert.
func ParseText(text string, observed time.Time) ([]Row, error) {
	matches := headerRe.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("acara guide header not found")
	}
	var rows []Row
	var model string
	brand := ""
	// ponytail: first guide line for a key/year wins. Later list codes (L20 vs L23) would overwrite.
	seen := map[string]struct{}{}
	for i, m := range matches {
		nextBrand := strings.TrimSpace(text[m[2]:m[3]])
		if nextBrand != brand {
			brand = nextBrand
			model = ""
		}
		years := yearRe.FindAllString(text[m[4]:m[5]], -1)
		end := len(text)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		var page []Row
		page, model = scanSection(text[m[1]:end], brand, model, years, observed)
		for _, row := range page {
			row.Cells = unseenCells(row, seen)
			if len(row.Cells) == 0 {
				continue
			}
			rows = append(rows, row)
		}
	}
	return rows, nil
}

func scanSection(section, brand, model string, years []string, observed time.Time) ([]Row, string) {
	n := 1 + len(years)
	var rows []Row
	i := 0
	for i < len(section) {
		pos, curr := nextCurrency(section[i:])
		if pos < 0 {
			break
		}
		desc := strings.TrimSpace(section[i : i+pos])
		rest := section[i+pos:]
		if curr == "USD" {
			rest = rest[len("u$s"):]
		} else {
			rest = rest[1:]
		}
		tokens, next, ok := takePrices(rest, n)
		if !ok {
			i += pos + 1
			continue
		}
		var trim string
		model, trim = splitDesc(desc, model)
		row := Row{Brand: brand, Model: model, Trim: trim, Currency: curr}
		row.Cells = cells(tokens, years, observed)
		if len(row.Cells) > 0 && strings.TrimSpace(trim) != "" {
			rows = append(rows, row)
		}
		i += pos
		if curr == "USD" {
			i += len("u$s")
		} else {
			i++
		}
		i += next
	}
	return rows, model
}

func unseenCells(row Row, seen map[string]struct{}) []Cell {
	key := ""
	if keys := row.Keys(); len(keys) > 0 {
		key = keys[0]
	}
	var out []Cell
	for _, c := range row.Cells {
		id := key + "/" + strconv.Itoa(c.Year)
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, c)
	}
	return out
}

func cells(tokens, years []string, observed time.Time) []Cell {
	var out []Cell
	issueYear := observed.Year()
	for i, tok := range tokens {
		amount, ok := parseAmount(tok)
		if !ok {
			continue
		}
		if i == 0 {
			if issueYear <= 0 || yearListed(years, issueYear) {
				continue
			}
			out = append(out, Cell{Year: issueYear, Price: amount})
			continue
		}
		year, err := strconv.Atoi(years[i-1])
		if err != nil || year <= 0 {
			continue
		}
		out = append(out, Cell{Year: year, Price: amount})
	}
	return out
}

func yearListed(years []string, year int) bool {
	s := strconv.Itoa(year)
	for _, y := range years {
		if y == s {
			return true
		}
	}
	return false
}

func nextCurrency(s string) (int, string) {
	usd := strings.Index(s, "u$s")
	dollar := -1
	for j := 0; j < len(s); j++ {
		if s[j] != '$' {
			continue
		}
		if usd >= 0 && j == usd+1 {
			continue
		}
		dollar = j
		break
	}
	switch {
	case usd >= 0 && (dollar < 0 || usd <= dollar):
		return usd, "USD"
	case dollar >= 0:
		return dollar, "ARS"
	default:
		return -1, ""
	}
}

func takePrices(s string, n int) ([]string, int, bool) {
	tokens := make([]string, 0, n)
	i := 0
	for len(tokens) < n {
		for i < len(s) && (s[i] == ' ' || s[i] == '\n' || s[i] == '\t' || s[i] == '\r') {
			i++
		}
		if i >= len(s) {
			return nil, 0, false
		}
		m := priceRe.FindString(s[i:])
		if m == "" {
			return nil, 0, false
		}
		tokens = append(tokens, m)
		i += len(m)
	}
	return tokens, i, true
}

func parseAmount(raw string) (int64, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "-" {
		return 0, false
	}
	if i := strings.IndexByte(raw, ','); i >= 0 {
		raw = raw[:i]
	}
	raw = strings.ReplaceAll(raw, ".", "")
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

func splitDesc(desc, model string) (string, string) {
	fields := strings.Fields(desc)
	if len(fields) == 0 {
		return model, ""
	}
	i := 0
	for i < len(fields) && !isTrimStart(fields[i]) {
		i++
	}
	if i == 0 {
		return model, strings.Join(fields, " ")
	}
	return strings.Join(fields[:i], " "), strings.Join(fields[i:], " ")
}

func isTrimStart(tok string) bool {
	if tok == "" || (tok[0] >= '0' && tok[0] <= '9') {
		return tok != ""
	}
	switch strings.ToUpper(tok) {
	case "HYBRID", "VARIANT", "SPORTBACK", "CABRIO", "COUPE", "COUPÉ", "SEDAN", "SEDÁN", "SW":
		return true
	default:
		return false
	}
}

// ExtractText pulls text from FlateDecode content streams (Tj / TJ).
func ExtractText(data []byte) string {
	var b strings.Builder
	for _, blob := range contentStreams(data) {
		if !bytes.Contains(blob, []byte("Tj")) && !bytes.Contains(blob, []byte("TJ")) {
			continue
		}
		for _, piece := range textPieces(blob) {
			b.WriteString(latin1(piece))
			b.WriteByte(' ')
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func contentStreams(data []byte) [][]byte {
	var out [][]byte
	i := 0
	for {
		s := bytes.Index(data[i:], []byte("stream"))
		if s < 0 {
			break
		}
		s += i
		start := s + len("stream")
		if bytes.HasPrefix(data[start:], []byte("\r\n")) {
			start += 2
		} else if start < len(data) && (data[start] == '\n' || data[start] == '\r') {
			start++
		}
		e := bytes.Index(data[start:], []byte("endstream"))
		if e < 0 {
			break
		}
		e += start
		dictAt := bytes.LastIndex(data[:s], []byte("<<"))
		header := []byte{}
		if dictAt >= 0 && s-dictAt < 8000 {
			header = data[dictAt:s]
		}
		raw := data[start:e]
		i = e + len("endstream")
		blob := raw
		if bytes.Contains(header, []byte("FlateDecode")) {
			var err error
			blob, err = inflate(raw)
			if err != nil {
				continue
			}
		}
		out = append(out, blob)
	}
	return out
}

func inflate(raw []byte) ([]byte, error) {
	raw = bytes.TrimRight(raw, "\r\n ")
	r, err := zlib.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

func textPieces(blob []byte) [][]byte {
	var out [][]byte
	for i := 0; i < len(blob); i++ {
		switch blob[i] {
		case '(':
			s, next := readLiteral(blob, i)
			j := skipSpace(blob, next)
			if bytes.HasPrefix(blob[j:], []byte("Tj")) {
				out = append(out, s)
				i = j + 1
			}
		case '[':
			end := bytes.IndexByte(blob[i:], ']')
			if end < 0 {
				continue
			}
			j := skipSpace(blob, i+end+1)
			if !bytes.HasPrefix(blob[j:], []byte("TJ")) {
				continue
			}
			out = append(out, literalsIn(blob[i+1:i+end])...)
			i = j + 1
		}
	}
	return out
}

func literalsIn(chunk []byte) [][]byte {
	var out [][]byte
	for i := 0; i < len(chunk); i++ {
		if chunk[i] != '(' {
			continue
		}
		s, next := readLiteral(chunk, i)
		out = append(out, s)
		if next <= i {
			break
		}
		i = next - 1
	}
	return out
}

func readLiteral(b []byte, open int) ([]byte, int) {
	var out []byte
	i := open + 1
	for i < len(b) {
		c := b[i]
		if c == '\\' && i+1 < len(b) {
			n := b[i+1]
			switch n {
			case 'n':
				out = append(out, '\n')
			case 'r':
				out = append(out, '\r')
			case 't':
				out = append(out, '\t')
			case '(', ')', '\\':
				out = append(out, n)
			default:
				if n >= '0' && n <= '7' {
					j := i + 1
					oct := 0
					for k := 0; k < 3 && j < len(b) && b[j] >= '0' && b[j] <= '7'; k++ {
						oct = oct*8 + int(b[j]-'0')
						j++
					}
					out = append(out, byte(oct))
					i = j
					continue
				}
				out = append(out, n)
			}
			i += 2
			continue
		}
		if c == ')' {
			return out, i + 1
		}
		out = append(out, c)
		i++
	}
	return out, len(b)
}

// WinAnsi bytes in this guide are not valid UTF-8 (í is 0xED).
func latin1(b []byte) string {
	r := make([]rune, len(b))
	for i, c := range b {
		r[i] = rune(c)
	}
	return string(r)
}

func skipSpace(b []byte, i int) int {
	for i < len(b) && (b[i] == ' ' || b[i] == '\n' || b[i] == '\r' || b[i] == '\t') {
		i++
	}
	return i
}
