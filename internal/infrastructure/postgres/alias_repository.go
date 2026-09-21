package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AliasRepository struct {
	pool *pgxpool.Pool
}

func NewAliasRepository(pool *pgxpool.Pool) *AliasRepository {
	return &AliasRepository{pool: pool}
}

func (r *AliasRepository) Resolve(ctx context.Context, source, raw string) (int, bool, error) {
	var trimID int
	err := r.pool.QueryRow(ctx, `
		SELECT trim_id FROM trim_aliases
		WHERE source = $1 AND raw_name = $2`, source, raw).Scan(&trimID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return trimID, true, nil
}
