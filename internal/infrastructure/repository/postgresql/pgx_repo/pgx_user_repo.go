package pgxrepo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepoPgx struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepoPgx {
	return &UserRepoPgx{pool: pool}
}

func (r *UserRepoPgx) CreateUser(ctx context.Context, id int64) error {
	sql := "INSERT INTO users(tg_id) VALUES($1)"
	_, err := r.pool.Exec(ctx, sql, id)

	return err
}

func (r *UserRepoPgx) DeleteUser(ctx context.Context, id int64) error {
	sql := "DELETE FROM users WHERE tg_id=$1"

	result, err := r.pool.Exec(ctx, sql, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *UserRepoPgx) GetAllUsers(ctx context.Context) ([]int64, error) {
	sql := "SELECT tg_id FROM users"

	rows, err := r.pool.Query(ctx, sql)
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

	return tgIDs, rows.Err()
}
