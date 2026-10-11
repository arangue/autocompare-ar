package cca

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/arangue/autocompare-ar/internal/infrastructure/ingestion/pdf/guide"
)

// Column clip origins for the October 2026 Autos list (0km, then 2025 down to 2012).
// A cell's moveto sits about 19px off the center, so snap within 40.
var columns = [...]struct {
	y    float64
	year int
}{
	{1439, 0},
	{1610, 2025},
	{1758, 2024},
	{1906, 2023},
	{2067, 2022},
	{2225, 2021},
	{2376, 2020},
	{2530, 2019},
	{2688, 2018},
	{2875, 2017},
	{3029, 2016},
	{3177, 2015},
	{3331, 2014},
	{3485, 2013},
	{3649, 2012},
}

// ParseFile reads a CCA Autos PDF into guide rows. Prices are whole US dollars.
// R10 is the bold face (brand, or a trim when the cell has prices) and R12 is
// the italic face (model). Both carry across pages.
func ParseFile(path string, observed time.Time) ([]guide.Row, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	rows, err := parsePDF(data, observed)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return rows, nil
}

func parsePDF(data []byte, observed time.Time) ([]guide.Row, error) {
	var bold, italic map[byte]rune
	var pages [][]byte
	for _, blob := range contentStreams(data) {
		if bytes.Contains(blob, []byte("beginbfrange")) || bytes.Contains(blob, []byte("beginbfchar")) {
			cm := parseCMap(blob)
			switch cm[1] {
			case 'A':
				bold = cm
			case 'M':
				italic = cm
			}
			continue
		}
		if bytes.Contains(blob, []byte("W n")) && bytes.Contains(blob, []byte("TJ")) {
			pages = append(pages, blob)
		}
	}
	if bold == nil || italic == nil {
		return nil, fmt.Errorf("cca font maps not found")
	}
	rows := rowsFromPages(pages, bold, italic, observed)
	if len(rows) == 0 {
		return nil, fmt.Errorf("no cca rows")
	}
	return rows, nil
}

type placed struct {
	x, y float64
	font string
	text string
}

func rowsFromPages(pages [][]byte, bold, italic map[byte]rune, observed time.Time) []guide.Row {
	var rows []guide.Row
	seen := map[string]struct{}{}
	var brand, model string
	issue := observed.Year()
	for _, page := range pages {
		groups := map[int][]placed{}
		var xs []int
		for _, c := range pageCells(page, bold, italic) {
			k := int(math.Round(c.x))
			if _, ok := groups[k]; !ok {
				xs = append(xs, k)
			}
			groups[k] = append(groups[k], c)
		}
		sort.Ints(xs)
		for _, x := range xs {
			items := groups[x]
			sort.Slice(items, func(i, j int) bool { return items[i].y < items[j].y })
			var label *placed
			var prices []guide.Cell
			used := map[int]bool{}
			for i := range items {
				c := &items[i]
				if c.y < 400 {
					if label == nil {
						label = c
					}
					continue
				}
				year, ok := yearAt(c.y)
				if !ok {
					continue
				}
				if year == 0 {
					if issue <= 0 || columnHas(issue) {
						continue
					}
					year = issue
				}
				n, ok := parsePrice(c.text)
				if !ok || used[year] {
					continue
				}
				used[year] = true
				prices = append(prices, guide.Cell{Year: year, Price: n})
			}
			if label == nil {
				continue
			}
			text := strings.Join(strings.Fields(label.text), " ")
			if text == "" {
				continue
			}
			if len(prices) == 0 {
				switch label.font {
				case "R12":
					model = text
				case "R10":
					brand = text
					model = ""
				}
				continue
			}
			if label.font != "R10" && label.font != "R12" {
				continue
			}
			row := guide.Row{Brand: brand, Model: model, Trim: text, Currency: "USD", Cells: prices}
			row.Cells = unseen(row, seen)
			if len(row.Cells) > 0 {
				rows = append(rows, row)
			}
		}
	}
	return rows
}

func unseen(row guide.Row, seen map[string]struct{}) []guide.Cell {
	key := ""
	if keys := row.Keys(); len(keys) > 0 {
		key = keys[0]
	}
	var out []guide.Cell
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

func yearAt(y float64) (int, bool) {
	best, dist := 0, 41.0
	for _, c := range columns {
		d := math.Abs(c.y - y)
		if d < dist {
			dist = d
			best = c.year
		}
	}
	if dist > 40 {
		return 0, false
	}
	return best, true
}

func columnHas(year int) bool {
	for _, c := range columns {
		if c.year == year {
			return true
		}
	}
	return false
}

// parsePrice reads a CCA amount. A comma plus three digits is a thousands
// separator (111,660 → 111660). A comma plus one or two digits is a fraction
// and is dropped (48410,0 → 48410).
func parsePrice(s string) (int64, bool) {
	s = strings.ReplaceAll(strings.TrimSpace(s), ".", "")
	s = strings.ReplaceAll(s, " ", "")
	if s == "" || s == "-" {
		return 0, false
	}
	if i := strings.LastIndexByte(s, ','); i >= 0 {
		whole, frac := strings.ReplaceAll(s[:i], ",", ""), s[i+1:]
		if whole == "" || !digits(whole) || !digits(frac) {
			return 0, false
		}
		switch {
		case len(frac) == 3:
			s = whole + frac
		case len(frac) <= 2:
			s = whole
		default:
			return 0, false
		}
	}
	if !digits(s) {
		return 0, false
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

func digits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

var (
	moveRe = regexp.MustCompile(`([\d.]+) ([\d.]+) m(?:\r\n|\n|\r)`)
	fontRe = regexp.MustCompile(`/(R\d+) [\d.]+ Tf`)
	bfRe   = regexp.MustCompile(`<([0-9A-Fa-f]+)>\s*<([0-9A-Fa-f]+)>\s*<([0-9A-Fa-f]+)>`)
	charRe = regexp.MustCompile(`<([0-9A-Fa-f]+)>\s*<([0-9A-Fa-f]{4})>`)
)

func pageCells(blob []byte, bold, italic map[byte]rune) []placed {
	var out []placed
	i := 0
	for {
		j := bytes.Index(blob[i:], []byte("W n"))
		if j < 0 {
			break
		}
		j += i
		headStart := j - 240
		if headStart < 0 {
			headStart = 0
		}
		matches := moveRe.FindAllSubmatch(blob[headStart:j], -1)
		i = j + 3
		if len(matches) == 0 {
			continue
		}
		last := matches[len(matches)-1]
		x, xerr := strconv.ParseFloat(string(last[1]), 64)
		y, yerr := strconv.ParseFloat(string(last[2]), 64)
		if xerr != nil || yerr != nil {
			continue
		}
		end := j + 2000
		if end > len(blob) {
			end = len(blob)
		}
		window := blob[j:end]
		fm := fontRe.FindSubmatch(window)
		if fm == nil {
			continue
		}
		rel := bytes.Index(window, fm[0])
		bracket := bytes.IndexByte(window[rel:], '[')
		if bracket < 0 {
			continue
		}
		bracket += rel
		closeAt := bytes.IndexByte(window[bracket:], ']')
		if closeAt < 0 {
			continue
		}
		after := skipSpace(window, bracket+closeAt+1)
		if !bytes.HasPrefix(window[after:], []byte("TJ")) {
			continue
		}
		raw := concatLiterals(window[bracket+1 : bracket+closeAt])
		cm := map[byte]rune(nil)
		font := string(fm[1])
		switch font {
		case "R10":
			cm = bold
		case "R12":
			cm = italic
		}
		text := strings.TrimSpace(applyCMap(raw, cm))
		if text == "" {
			continue
		}
		out = append(out, placed{x: x, y: y, font: font, text: text})
	}
	return out
}

func parseCMap(blob []byte) map[byte]rune {
	text := string(blob)
	m := map[byte]rune{}
	for _, block := range between(text, "beginbfrange", "endbfrange") {
		for _, sub := range bfRe.FindAllStringSubmatch(block, -1) {
			start := hex(sub[1])
			end := hex(sub[2])
			code := hex(sub[3])
			for i := 0; start+i <= end; i++ {
				src := start + i
				if src < 0 || src > 255 {
					continue
				}
				m[byte(src)] = rune(code + i)
			}
		}
	}
	for _, block := range between(text, "beginbfchar", "endbfchar") {
		for _, sub := range charRe.FindAllStringSubmatch(block, -1) {
			src := hex(sub[1])
			if src < 0 || src > 255 {
				continue
			}
			m[byte(src)] = rune(hex(sub[2]))
		}
	}
	return m
}

func between(s, start, end string) []string {
	var out []string
	for {
		i := strings.Index(s, start)
		if i < 0 {
			return out
		}
		s = s[i+len(start):]
		j := strings.Index(s, end)
		if j < 0 {
			return out
		}
		out = append(out, s[:j])
		s = s[j+len(end):]
	}
}

func hex(s string) int {
	n, _ := strconv.ParseInt(s, 16, 32)
	return int(n)
}

func applyCMap(raw []byte, cm map[byte]rune) string {
	if cm == nil {
		r := make([]rune, len(raw))
		for i, b := range raw {
			r[i] = rune(b)
		}
		return string(r)
	}
	r := make([]rune, len(raw))
	for i, b := range raw {
		ch, ok := cm[b]
		if !ok {
			r[i] = '?'
			continue
		}
		r[i] = ch
	}
	return string(r)
}

func concatLiterals(chunk []byte) []byte {
	var out []byte
	for i := 0; i < len(chunk); i++ {
		if chunk[i] != '(' {
			continue
		}
		s, next := readLiteral(chunk, i)
		out = append(out, s...)
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

func skipSpace(b []byte, i int) int {
	for i < len(b) && (b[i] == ' ' || b[i] == '\n' || b[i] == '\r' || b[i] == '\t') {
		i++
	}
	return i
}
