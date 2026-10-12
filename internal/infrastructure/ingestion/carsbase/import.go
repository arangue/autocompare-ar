package carsbase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "modernc.org/sqlite"
)

// Stats is the result of one import.
type Stats struct {
	Sheets  int
	Dropped int
	Links   int
	Trims   int
}

// Load replaces equipment_sheets with the sqlite file and relinks catalog
// trims. dryRun reports what would be written and leaves Postgres unchanged.
func Load(ctx context.Context, pool *pgxpool.Pool, sqlitePath string, dryRun bool) (Stats, error) {
	sheets, dropped, err := readSQLite(sqlitePath)
	if err != nil {
		return Stats{}, err
	}
	trims, err := loadTrims(ctx, pool)
	if err != nil {
		return Stats{}, err
	}
	links := MatchAll(trims, sheets)
	stats := Stats{Sheets: len(sheets), Dropped: dropped, Links: len(links), Trims: linkedTrims(links)}
	if dryRun {
		return stats, nil
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return Stats{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `DELETE FROM equipment_sheets`); err != nil {
		return Stats{}, err
	}
	if err := insertSheets(ctx, tx, sheets); err != nil {
		return Stats{}, err
	}
	if err := insertLinks(ctx, tx, links); err != nil {
		return Stats{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Stats{}, err
	}
	return stats, nil
}

func linkedTrims(links []Link) int {
	seen := map[int]bool{}
	for _, link := range links {
		seen[link.TrimID] = true
	}
	return len(seen)
}

func readSQLite(path string) ([]Sheet, int, error) {
	db, err := sql.Open("sqlite", "file:"+path+"?immutable=1")
	if err != nil {
		return nil, 0, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT s.url, COALESCE(NULLIF(m.name, ''), s.brand), COALESCE(NULLIF(m.years, ''), s.model),
		       s.trim_name, s.year, s.detalles_json
		FROM specs s
		LEFT JOIN models m ON m.brand = s.brand AND m.model_slug = s.model`)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	seen := map[string]bool{}
	var sheets []Sheet
	dropped := 0
	for rows.Next() {
		var url, brand, model, title, yearText, detalles string
		if err := rows.Scan(&url, &brand, &model, &title, &yearText, &detalles); err != nil {
			return nil, 0, err
		}
		year, err := strconv.Atoi(yearText)
		if err != nil || url == "" || seen[url] {
			dropped++
			continue
		}
		from, to, open, ok := TitleSpan(title)
		if !ok {
			dropped++
			continue
		}
		parsed, err := ParseDetalles(detalles)
		if err != nil {
			dropped++
			continue
		}
		parsed.URL = url
		parsed.Brand = brand
		parsed.Model = model
		parsed.Title = title
		parsed.Year = year
		parsed.YearFrom = from
		if !open {
			parsed.YearTo = &to
		}
		seen[url] = true
		sheets = append(sheets, parsed)
	}
	return sheets, dropped, rows.Err()
}

func loadTrims(ctx context.Context, pool *pgxpool.Pool) ([]TrimRef, error) {
	rows, err := pool.Query(ctx, `
		SELECT t.id, b.name, m.name, t.name, g.year_from, COALESCE(g.year_to, 0),
		       COALESCE(t.year_from, 0), COALESCE(t.year_to, 0)
		FROM trims t
		JOIN generations g ON g.id = t.generation_id
		JOIN models m ON m.id = g.model_id
		JOIN brands b ON b.id = m.brand_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trims []TrimRef
	for rows.Next() {
		var trim TrimRef
		if err := rows.Scan(&trim.ID, &trim.Brand, &trim.Model, &trim.Name, &trim.GenFrom, &trim.GenTo, &trim.YearFrom, &trim.YearTo); err != nil {
			return nil, err
		}
		trims = append(trims, trim)
	}
	return trims, rows.Err()
}

func insertSheets(ctx context.Context, tx pgx.Tx, sheets []Sheet) error {
	const q = `INSERT INTO equipment_sheets (
		url, brand_name, model_name, title, year, year_from, year_to,
		engine, displacement_cc, horsepower, transmission, fuel, doors, seats,
		consumption_city, consumption_highway, features
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17::jsonb)`
	for _, sheet := range sheets {
		raw, err := json.Marshal(sheet.Features)
		if err != nil {
			return err
		}
		if sheet.Features == nil {
			raw = []byte("[]")
		}
		if _, err := tx.Exec(ctx, q,
			sheet.URL, sheet.Brand, sheet.Model, sheet.Title, sheet.Year, sheet.YearFrom, sheet.YearTo,
			sheet.Engine, sheet.DisplacementCC, sheet.Horsepower, sheet.Transmission, sheet.Fuel,
			sheet.Doors, sheet.Seats, sheet.ConsumptionCity, sheet.ConsumptionHighway, string(raw),
		); err != nil {
			return fmt.Errorf("insert %s: %w", sheet.URL, err)
		}
	}
	return nil
}

func insertLinks(ctx context.Context, tx pgx.Tx, links []Link) error {
	for _, link := range links {
		if _, err := tx.Exec(ctx, `
			INSERT INTO trim_equipment (trim_id, year, sheet_url) VALUES ($1, $2, $3)`,
			link.TrimID, link.Year, link.URL,
		); err != nil {
			return fmt.Errorf("link trim %d year %d: %w", link.TrimID, link.Year, err)
		}
	}
	return nil
}
