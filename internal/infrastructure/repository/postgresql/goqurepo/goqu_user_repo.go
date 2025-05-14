package goqurepo

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type UserRepoGoqu struct {
	pool *pgxpool.Pool
	db   *goqu.Database
}

func NewUserRepoGoqu(pool *pgxpool.Pool) *UserRepoGoqu {
	sqlDB := stdlib.OpenDBFromPool(pool)
	db := goqu.New("postgres", sqlDB)

	return &UserRepoGoqu{
		pool: pool,
		db:   db,
	}
}

func (r *UserRepoGoqu) CreateUser(ctx context.Context, id int64) error {
	ds := r.db.Insert("users").
		Cols("tg_id").
		Vals(goqu.Vals{id})

	sql, args, err := ds.ToSQL()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, sql, args...)

	return err
}

func (r *UserRepoGoqu) DeleteUser(ctx context.Context, id int64) error {
	ds := r.db.Delete("users").Where(goqu.Ex{"tg_id": id})

	sql, args, err := ds.ToSQL()
	if err != nil {
		return err
	}

	result, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *UserRepoGoqu) GetAllUsers(ctx context.Context) ([]int64, error) {
	ds := r.db.From("users").Select("tg_id")

	sql, args, err := ds.ToSQL()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tgIDs []int64

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		tgIDs = append(tgIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tgIDs, nil
}
