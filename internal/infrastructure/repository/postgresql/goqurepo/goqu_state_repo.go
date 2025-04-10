package goqurepo

import (
	"LinkTracker/internal/domain"
	"context"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StateRepoGoqu struct {
	pool *pgxpool.Pool
	db   *goqu.Database
}

func NewStateRepoGoqu(pool *pgxpool.Pool) *StateRepoGoqu {
	sqlDB := stdlib.OpenDBFromPool(pool)
	db := goqu.New("postgres", sqlDB)

	return &StateRepoGoqu{
		pool: pool,
		db:   db,
	}
}

func (r *StateRepoGoqu) CreateState(ctx context.Context, tgID int64, state int) error {
	ds := r.db.Insert("states").
		Rows(goqu.Record{"tg_id": tgID, "state": state}).
		OnConflict(goqu.DoUpdate("tg_id", goqu.Record{"state": state}))

	sql, args, err := ds.ToSQL()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *StateRepoGoqu) DeleteState(ctx context.Context, tgID int64) error {
	ds := r.db.Delete("states").Where(goqu.Ex{"tg_id": tgID})

	sql, args, err := ds.ToSQL()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *StateRepoGoqu) GetState(ctx context.Context, tgID int64) (int, domain.Link, error) {
	ds := r.db.From("states").Select("state", "url", "tags", "filters").Where(goqu.Ex{"tg_id": tgID})

	sql, args, err := ds.ToSQL()
	if err != nil {
		return -1, domain.Link{}, err
	}

	var (
		state   int
		url     pgtype.Text
		tags    pgtype.Array[string]
		filters pgtype.Array[string]
	)

	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&state, &url, &tags, &filters); err != nil {
		return -1, domain.Link{}, err
	}

	link := domain.Link{
		URL:     "",
		Tags:    []string{},
		Filters: []string{},
	}
	if url.Valid {
		link.URL = url.String
	}
	if tags.Valid {
		link.Tags = tags.Elements
	}
	if filters.Valid {
		link.Filters = filters.Elements
	}

	return state, link, nil
}

func (r *StateRepoGoqu) UpdateState(ctx context.Context, tgID int64, state int, link *domain.Link) error {
	ds := r.db.Update("states").
		Set(goqu.Record{
			"state":   state,
			"url":     link.URL,
			"tags":    stringArrayToPostgres(link.Tags),
			"filters": stringArrayToPostgres(link.Filters),
		}).
		Where(goqu.Ex{"tg_id": tgID})

	sql, args, err := ds.ToSQL()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}
