package pgxrepo

import (
	"LinkTracker/internal/domain"
	"context"
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
	sql := "INSERT INTO states (tg_id, state) VALUES ($1, $2)"

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

func (r *StateRepoPgx) GetState(ctx context.Context, tgID int64) (int, error) {
	sql := "SELECT state FROM states WHERE tg_id = $1"
	row := r.pool.QueryRow(ctx, sql, tgID)

	var state int

	err := row.Scan(&state)
	if err != nil {
		return 0, err
	}

	return state, nil
}

func (r *StateRepoPgx) UpdateState(ctx context.Context, tgID int64, state int) error {
	sql := "UPDATE states SET state = $1 WHERE tg_id = $2"
	_, err := r.pool.Exec(ctx, sql, state, tgID)
	if err != nil {
		return err
	}

	return nil
}

func (r *StateRepoPgx) UpdateURL(ctx context.Context, tgID int64, linkURL string) error {
	sql := "UPDATE states SET url = $1 WHERE tg_id = $2"
	_, err := r.pool.Exec(ctx, sql, linkURL, tgID)
	if err != nil {
		return err
	}
	return nil
}

func (r *StateRepoPgx) UpdateTags(ctx context.Context, tgID int64, tags []string) error {
	sql := "UPDATE states SET tags = $1 WHERE tg_id = $2"
	_, err := r.pool.Exec(ctx, sql, tags, tgID)
	if err != nil {
		return err
	}
	return nil
}

func (r *StateRepoPgx) UpdateFilters(ctx context.Context, tgID int64, filters []string) error {
	sql := "UPDATE states SET filters = $1 WHERE tg_id = $2"
	_, err := r.pool.Exec(ctx, sql, filters, tgID)
	if err != nil {
		return err
	}
	return nil
}

func (r *StateRepoPgx) GetStateLink(ctx context.Context, tgID int64) (domain.Link, error) {
	sql := "SELECT url,tags,filters FROM states WHERE tg_id = $1"
	row := r.pool.QueryRow(ctx, sql, tgID)
	var link domain.Link
	err := row.Scan(&link.URL, &link.Tags, &link.Filters)
	if err != nil {
		return domain.Link{}, err
	}

	return link, nil
}
