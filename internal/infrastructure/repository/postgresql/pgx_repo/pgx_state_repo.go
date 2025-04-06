package pgxrepo

import (
	"context"

	"LinkTracker/internal/domain"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StateRepoPgx struct {
	pool *pgxpool.Pool
}

func NewStateRepoPgx(pool *pgxpool.Pool) *StateRepoPgx {
	return &StateRepoPgx{
		pool: pool,
	}
}

func (r *StateRepoPgx) CreateState(ctx context.Context, tgID int64, state int) error {
	sql := "INSERT INTO states (tg_id, state) VALUES ($1, $2) ON CONFLICT(tg_id) DO UPDATE SET state = $2"

	_, err := r.pool.Exec(ctx, sql, tgID, state)
	if err != nil {
		return err
	}

	return nil
}

func (r *StateRepoPgx) DeleteState(ctx context.Context, tgID int64) error {
	sql := "DELETE FROM states  WHERE tg_id = $1"

	_, err := r.pool.Exec(ctx, sql, tgID)
	if err != nil {
		return err
	}

	return nil
}

func (r *StateRepoPgx) GetState(ctx context.Context, tgID int64) (int, domain.Link, error) {
	sqlSelect := "SELECT state,url,tags,filters FROM states WHERE tg_id = $1"
	row := r.pool.QueryRow(ctx, sqlSelect, tgID)

	var (
		state   int
		url     pgtype.Text
		tags    pgtype.Array[string]
		filters pgtype.Array[string]
	)

	if err := row.Scan(&state, &url, &tags, &filters); err != nil {
		return -1, domain.Link{}, err
	}

	link := domain.Link{
		URL:     "",         // значение по умолчанию
		Tags:    []string{}, // значение по умолчанию
		Filters: []string{}, // значение по умолчанию
	}

	// Если url не null, используем его
	if url.Valid {
		link.URL = url.String
	}
	// Если tags или filters равны nil, оставляем пустой срез (по умолчанию)
	if tags.Valid {
		link.Tags = tags.Elements
	}
	if filters.Valid {
		link.Filters = filters.Elements
	}

	return state, link, nil
}

func (r *StateRepoPgx) UpdateState(ctx context.Context, tgID int64, state int, link domain.Link) error {
	url := link.URL
	tags := link.Tags
	filters := link.Filters

	sql := "UPDATE states SET state = $1, url=$2,tags = $3, filters= $4 WHERE tg_id = $5"
	_, err := r.pool.Exec(ctx, sql, state, url, tags, filters, tgID)
	if err != nil {
		return err
	}

	return nil
}
