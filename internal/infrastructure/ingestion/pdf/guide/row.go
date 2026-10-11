package guide

import "strings"

// Row is one line from a price guide. Cells are whole currency units.
type Row struct {
	Brand    string
	Model    string
	Trim     string
	Currency string
	Cells    []Cell
}

// Cell is one year. The 0km column uses the guide issue year when that year is not already a column.
type Cell struct {
	Year  int
	Price int64
}

// Keys are alias candidates before normalize.Title. Parentheticals such as (170cv) are dropped.
func (r Row) Keys() []string {
	trim := stripParens(r.Trim)
	model := strings.TrimSpace(r.Model)
	brand := strings.TrimSpace(r.Brand)
	var keys []string
	add := func(s string) {
		s = strings.Join(strings.Fields(s), " ")
		if s == "" {
			return
		}
		for _, k := range keys {
			if k == s {
				return
			}
		}
		keys = append(keys, s)
	}
	if model != "" && !strings.HasPrefix(strings.ToUpper(trim)+" ", strings.ToUpper(model)+" ") {
		add(brand + " " + model + " " + trim)
		add(model + " " + trim)
	} else {
		add(brand + " " + trim)
	}
	add(trim)
	return keys
}

func stripParens(s string) string {
	for {
		open := strings.IndexByte(s, '(')
		if open < 0 {
			break
		}
		close := strings.IndexByte(s[open:], ')')
		if close < 0 {
			break
		}
		s = s[:open] + s[open+close+1:]
	}
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}
