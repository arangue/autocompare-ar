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

// Load replaces equipment_sheets with the sqlite file and upserts a catalog
// trim for every title. dryRun reports what would be written and leaves
// Postgres unchanged.
func Load(ctx context.Context, pool *pgxpool.Pool, sqlitePath string, dryRun bool) (Stats, error) {
	sheets, dropped, err := readSQLite(sqlitePath)
	if err != nil {
		return Stats{}, err
	}
	catalog := BuildCatalog(sheets)
	links := 0
	for _, trim := range catalog {
		links += len(trim.Years)
	}
	stats := Stats{Sheets: len(sheets), Dropped: dropped, Links: links, Trims: len(catalog)}
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
	if err := insertCatalog(ctx, tx, catalog); err != nil {
		return Stats{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Stats{}, err
	}
	return stats, nil
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

func insertCatalog(ctx context.Context, tx pgx.Tx, trims []CatalogTrim) error {
	type span struct {
		from int
		to   *int
	}
	spans := map[string]*span{}
	for _, trim := range trims {
		key := trim.Brand + "\x00" + trim.Model + "\x00" + trim.Generation
		s := spans[key]
		if s == nil {
			s = &span{from: trim.YearFrom, to: trim.YearTo}
			spans[key] = s
			continue
		}
		if trim.YearFrom < s.from {
			s.from = trim.YearFrom
		}
		if trim.YearTo == nil || s.to == nil {
			s.to = nil
			continue
		}
		if *trim.YearTo > *s.to {
			v := *trim.YearTo
			s.to = &v
		}
	}

	brands := map[string]int{}
	models := map[string]int{}
	gens := map[string]int{}
	for _, trim := range trims {
		brandID, err := cachedID(brands, trim.Brand, func() (int, error) {
			return upsertID(ctx, tx, `
				INSERT INTO brands (name) VALUES ($1)
				ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
				RETURNING id`, trim.Brand)
		})
		if err != nil {
			return err
		}
		modelKey := trim.Brand + "\x00" + trim.Model
		modelID, err := cachedID(models, modelKey, func() (int, error) {
			return upsertID(ctx, tx, `
				INSERT INTO models (brand_id, name) VALUES ($1, $2)
				ON CONFLICT (brand_id, name) DO UPDATE SET name = EXCLUDED.name
				RETURNING id`, brandID, trim.Model)
		})
		if err != nil {
			return err
		}
		genKey := modelKey + "\x00" + trim.Generation
		genSpan := spans[genKey]
		genID, err := cachedID(gens, genKey, func() (int, error) {
			return upsertID(ctx, tx, `
				INSERT INTO generations (model_id, name, year_from, year_to) VALUES ($1, $2, $3, $4)
				ON CONFLICT (model_id, name) DO UPDATE SET
					year_from = EXCLUDED.year_from,
					year_to = EXCLUDED.year_to
				RETURNING id`, modelID, trim.Generation, genSpan.from, genSpan.to)
		})
		if err != nil {
			return err
		}
		trimID, err := upsertID(ctx, tx, `
			INSERT INTO trims (generation_id, name, year_from, year_to) VALUES ($1, $2, $3, $4)
			ON CONFLICT (generation_id, name) DO UPDATE SET
				year_from = EXCLUDED.year_from,
				year_to = EXCLUDED.year_to
			RETURNING id`, genID, trim.Name, trim.YearFrom, trim.YearTo)
		if err != nil {
			return fmt.Errorf("trim %s %s %s: %w", trim.Brand, trim.Model, trim.Name, err)
		}
		for year, url := range trim.Years {
			if _, err := tx.Exec(ctx, `
				INSERT INTO trim_equipment (trim_id, year, sheet_url) VALUES ($1, $2, $3)`,
				trimID, year, url,
			); err != nil {
				return fmt.Errorf("link %s %d: %w", trim.Name, year, err)
			}
		}
	}
	return nil
}

func cachedID(cache map[string]int, key string, load func() (int, error)) (int, error) {
	if id, ok := cache[key]; ok {
		return id, nil
	}
	id, err := load()
	if err != nil {
		return 0, err
	}
	cache[key] = id
	return id, nil
}

func upsertID(ctx context.Context, tx pgx.Tx, query string, args ...any) (int, error) {
	var id int
	err := tx.QueryRow(ctx, query, args...).Scan(&id)
	return id, err
}

func insertSheets(ctx context.Context, tx pgx.Tx, sheets []Sheet) error {
	const q = `INSERT INTO equipment_sheets (
		url, brand_name, model_name, title, year, year_from, year_to,
		engine, displacement_cc, horsepower, transmission, fuel, doors, seats,
		consumption_city, consumption_highway, consumption_mixed,
		torque_nm, cylinders, top_speed_kmh, traction, features
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22::jsonb)`
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
			sheet.Doors, sheet.Seats, sheet.ConsumptionCity, sheet.ConsumptionHighway, sheet.ConsumptionMixed,
			sheet.TorqueNm, sheet.Cylinders, sheet.TopSpeedKmh, sheet.Traction, string(raw),
		); err != nil {
			return fmt.Errorf("insert %s: %w", sheet.URL, err)
		}
	}
	return nil
}
