package pgxrepo

import (
	"context"
	"errors"
	"time"

	"LinkTracker/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LinkRepoPgx struct {
	pool *pgxpool.Pool
}

func NewLinkRepo(pool *pgxpool.Pool) *LinkRepoPgx {
	return &LinkRepoPgx{pool: pool}
}

func (r *LinkRepoPgx) GetUsersByLink(ctx context.Context, linkID int64) ([]int64, error) {
	sql := "SELECT tg_id FROM tracks  WHERE url_id = $1"

	rows, err := r.pool.Query(ctx, sql, linkID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tgIDs []int64

	for rows.Next() {
		var tgID int64

		err = rows.Scan(&tgID)
		if err != nil {
			return nil, err
		}

		tgIDs = append(tgIDs, tgID)
	}

	return tgIDs, rows.Err()
}

func (r *LinkRepoPgx) GetUserLinks(ctx context.Context, id int64) ([]domain.Link, error) {
	sql := `
        SELECT t.url_id, u.url, t.filters, t.tags
        FROM tracks t
        JOIN urls u ON t.url_id = u.id
        WHERE t.tg_id = $1
    `

	rows, err := r.pool.Query(ctx, sql, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []domain.Link

	for rows.Next() {
		var link domain.Link

		err = rows.Scan(&link.ID, &link.URL, &link.Filters, &link.Tags)
		if err != nil {
			return nil, err
		}

		links = append(links, link)
	}

	return links, rows.Err()
}

func (r *LinkRepoPgx) AddLink(ctx context.Context, id int64, link *domain.Link) (domain.Link, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Link{}, err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback(ctx))
		}
	}()

	var urlID int64

	sqlInsertLink := "INSERT INTO urls(url, last_update) VALUES($1, $2) RETURNING id"

	err = tx.QueryRow(ctx, sqlInsertLink, link.URL, time.Now().UTC()).Scan(&urlID)
	if err != nil {
		return domain.Link{}, err
	}

	sqlInsertTrack := "INSERT INTO tracks(tg_id, url_id, filters, tags) VALUES($1, $2, $3, $4)"

	_, err = tx.Exec(ctx, sqlInsertTrack, id, urlID, link.Filters, link.Tags)
	if err != nil {
		return domain.Link{}, err
	}

	link.ID = urlID

	return *link, tx.Commit(ctx)
}

func (r *LinkRepoPgx) DeleteLink(ctx context.Context, id int64, link *domain.Link) (domain.Link, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Link{}, err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback(ctx))
		}
	}()

	var (
		urlID      int64
		lastUpdate time.Time
		filters    []string
		tags       []string
	)

	sqlSelectUrlsAndTracks := `
		SELECT u.id, u.last_update, t.filters, t.tags 
		FROM urls u JOIN tracks t ON u.id = t.url_id 
		WHERE u.url = $1 AND t.tg_id = $2
	`

	err = tx.QueryRow(ctx, sqlSelectUrlsAndTracks, link.URL, id).Scan(&urlID, &lastUpdate, &filters, &tags)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Link{}, domain.ErrLinkNotExist{}
		}

		return domain.Link{}, err
	}

	sqlDeleteTrack := `DELETE FROM tracks WHERE tg_id = $1 AND url_id = $2`

	result, err := tx.Exec(ctx, sqlDeleteTrack, id, urlID)
	if err != nil {
		return domain.Link{}, err
	}

	if result.RowsAffected() == 0 {
		return domain.Link{}, pgx.ErrNoRows
	}

	sqlSelectCount := "SELECT COUNT(*) FROM tracks WHERE url_id = $1"

	var count int

	err = tx.QueryRow(ctx, sqlSelectCount, urlID).Scan(&count)
	if err != nil {
		return domain.Link{}, err
	}

	if count == 0 {
		sqlDeleteURL := "DELETE FROM urls WHERE id = $1"

		_, err = tx.Exec(ctx, sqlDeleteURL, urlID)
		if err != nil {
			return domain.Link{}, err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return domain.Link{}, err
	}

	return domain.Link{URL: link.URL, ID: urlID, Filters: filters, Tags: tags}, nil
}

func (r *LinkRepoPgx) GetAllLinks(ctx context.Context) ([]domain.Link, error) {
	sql := "SELECT id,url,last_update FROM urls"

	rows, err := r.pool.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []domain.Link

	for rows.Next() {
		var link domain.Link

		err = rows.Scan(&link.ID, &link.URL, &link.LastUpdated)
		if err != nil {
			return nil, err
		}

		links = append(links, link)
	}

	return links, rows.Err()
}

func (r *LinkRepoPgx) GetLinksAfter(ctx context.Context, lastUpdate time.Time, limit int64) ([]domain.Link, error) {
	sql := `
		SELECT id, url, last_update
		FROM urls
		WHERE last_update > $1
		ORDER BY last_update
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, sql, lastUpdate, limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var links []domain.Link

	for rows.Next() {
		var link domain.Link
		if err := rows.Scan(&link.ID, &link.URL, &link.LastUpdated); err != nil {
			return nil, err
		}

		links = append(links, link)
	}

	return links, rows.Err()
}

func (r *LinkRepoPgx) UpdateLink(ctx context.Context, tgID int64, link *domain.Link) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback(ctx))
		}
	}()

	getURLIDsql := "SELECT id FROM urls WHERE url = $1"

	var urlID int64

	err = tx.QueryRow(ctx, getURLIDsql, link.URL).Scan(&urlID)
	if err != nil {
		return err
	}

	sql := "UPDATE tracks SET tags = $1, filters = $2 WHERE tg_id = $3 AND url_id = $4"

	_, err = tx.Exec(ctx, sql, link.Tags, link.Filters, tgID, urlID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *LinkRepoPgx) UpdateTimeLink(ctx context.Context, lastUpdate time.Time, id int64) error {
	sql := "UPDATE urls SET last_update = $1 WHERE id = $2"
	_, err := r.pool.Exec(ctx, sql, lastUpdate, id)

	return err
}
