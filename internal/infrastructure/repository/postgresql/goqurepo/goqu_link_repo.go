package goqurepo

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"LinkTracker/internal/domain"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// LinkRepoGoqu реализует методы работы с БД.
type LinkRepoGoqu struct {
	pool *pgxpool.Pool
	db   *goqu.Database
}

// NewLinkRepoGoqu создаёт новый репозиторий.
func NewLinkRepoGoqu(pool *pgxpool.Pool) *LinkRepoGoqu {
	sqlDB := stdlib.OpenDBFromPool(pool)
	db := goqu.New("postgres", sqlDB)

	return &LinkRepoGoqu{
		pool: pool,
		db:   db,
	}
}

// GetUsersByLink возвращает идентификаторы пользователей, отслеживающих заданную ссылку.
func (r *LinkRepoGoqu) GetUsersByLink(ctx context.Context, linkID int64) ([]int64, error) {
	ds := r.db.From("tracks").Select("tg_id").Where(goqu.Ex{"url_id": linkID})

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
		var tgID int64
		if err := rows.Scan(&tgID); err != nil {
			return nil, err
		}

		tgIDs = append(tgIDs, tgID)
	}

	return tgIDs, rows.Err()
}

// GetUserLinks возвращает все ссылки пользователя с их связанными фильтрами и тегами.
func (r *LinkRepoGoqu) GetUserLinks(ctx context.Context, id int64) ([]domain.Link, error) {
	// Формируем запрос: SELECT t.url_id, u.url, t.filters, t.tags FROM tracks t JOIN urls u ON t.url_id = u.id WHERE t.tg_id = ?
	ds := r.db.From("tracks").
		Join(goqu.I("urls"), goqu.On(goqu.Ex{"tracks.url_id": goqu.I("urls.id")})).
		Select("tracks.url_id", "urls.url", "tracks.filters", "tracks.tags").
		Where(goqu.Ex{"tracks.tg_id": id})

	sql, args, err := ds.ToSQL()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []domain.Link

	for rows.Next() {
		var link domain.Link
		// При условии, что pgx может корректно сканировать TEXT ARRAY в []string
		if err := rows.Scan(&link.ID, &link.URL, &link.Filters, &link.Tags); err != nil {
			return nil, err
		}

		links = append(links, link)
	}

	return links, rows.Err()
}

// AddLink добавляет новую ссылку, вставляя её в таблицу urls и создавая трек в таблице tracks.
func (r *LinkRepoGoqu) AddLink(ctx context.Context, id int64, link *domain.Link) (domain.Link, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Link{}, err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback(ctx))
		}
	}()

	// Вставляем запись в таблицу urls с возвратом id
	dsInsertURL := r.db.Insert("urls").
		Cols("url", "last_update").
		Vals(goqu.Vals{link.URL, time.Now().UTC()}).
		Returning("id")

	sqlURL, argsURL, err := dsInsertURL.ToSQL()
	if err != nil {
		return domain.Link{}, err
	}

	var urlID int64
	if err := tx.QueryRow(ctx, sqlURL, argsURL...).Scan(&urlID); err != nil {
		return domain.Link{}, err
	}

	// Вставляем трек для пользователя и созданного url
	dsInsertTrack := r.db.Insert("tracks").
		Cols("tg_id", "url_id", "filters", "tags").
		Vals(goqu.Vals{id, urlID, stringArrayToPostgres(link.Filters), stringArrayToPostgres(link.Tags)})

	sqlTrack, argsTrack, err := dsInsertTrack.ToSQL()
	if err != nil {
		return domain.Link{}, err
	}

	if _, err = tx.Exec(ctx, sqlTrack, argsTrack...); err != nil {
		return domain.Link{}, err
	}

	link.ID = urlID

	return *link, tx.Commit(ctx)
}

// DeleteLink удаляет трек ссылки пользователя. Если для ссылки не осталось треков, удаляет запись из urls.
func (r *LinkRepoGoqu) DeleteLink(ctx context.Context, id int64, link *domain.Link) (domain.Link, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Link{}, err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback(ctx))
		}
	}()

	linkInfo, err := r.selectLinkInfo(ctx, tx, id, link)
	if err != nil {
		return domain.Link{}, err
	}

	if err := r.deleteTrack(ctx, tx, id, linkInfo.ID); err != nil {
		return domain.Link{}, err
	}

	if err := r.maybeDeleteURL(ctx, tx, linkInfo.ID); err != nil {
		return domain.Link{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Link{}, err
	}

	return *linkInfo, nil
}
func (r *LinkRepoGoqu) selectLinkInfo(ctx context.Context, tx pgx.Tx, id int64, link *domain.Link) (*domain.Link, error) {
	ds := r.db.From("urls").Join(goqu.I("tracks"), goqu.On(goqu.Ex{"urls.id": goqu.I("tracks.url_id")})).
		Select("urls.id", "urls.last_update", "tracks.filters", "tracks.tags").
		Where(goqu.Ex{"urls.url": link.URL, "tracks.tg_id": id})

	sql, args, err := ds.ToSQL()
	if err != nil {
		return nil, err
	}

	var l domain.Link

	l.URL = link.URL
	if err := tx.QueryRow(ctx, sql, args...).Scan(&l.ID, &l.LastUpdated, &l.Filters, &l.Tags); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrLinkNotExist{}
		}

		return nil, err
	}

	return &l, nil
}

func (r *LinkRepoGoqu) deleteTrack(ctx context.Context, tx pgx.Tx, tgID, urlID int64) error {
	ds := r.db.Delete("tracks").Where(goqu.Ex{"tg_id": tgID, "url_id": urlID})

	sql, args, err := ds.ToSQL()
	if err != nil {
		return err
	}

	result, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *LinkRepoGoqu) maybeDeleteURL(ctx context.Context, tx pgx.Tx, urlID int64) error {
	dsCount := r.db.From("tracks").Select(goqu.COUNT("*")).Where(goqu.Ex{"url_id": urlID})

	sqlCount, argsCount, err := dsCount.ToSQL()
	if err != nil {
		return err
	}

	var count int
	if err := tx.QueryRow(ctx, sqlCount, argsCount...).Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	dsDelete := r.db.Delete("urls").Where(goqu.Ex{"id": urlID})

	sqlDelete, argsDelete, err := dsDelete.ToSQL()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, sqlDelete, argsDelete...)

	return err
}

// GetAllLinks возвращает все записи из таблицы urls.
func (r *LinkRepoGoqu) GetAllLinks(ctx context.Context) ([]domain.Link, error) {
	ds := r.db.From("urls").Select("id", "url", "last_update")

	sql, args, err := ds.ToSQL()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
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

// GetLinksAfter возвращает записи из таблицы urls с last_update > заданного значения.
func (r *LinkRepoGoqu) GetLinksAfter(ctx context.Context, lastUpdate time.Time, limit int64) ([]domain.Link, error) {
	ds := r.db.From("urls").
		Select("id", "url", "last_update").
		Where(goqu.C("last_update").Gt(lastUpdate)).
		Order(goqu.C("last_update").Asc()).
		Limit(uint(limit)) //nolint // integer overflow conversion int64 -> uint (gosec) it is impossible

	sql, args, err := ds.ToSQL()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
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

// UpdateLink обновляет теги и фильтры для трека пользователя.
func (r *LinkRepoGoqu) UpdateLink(ctx context.Context, tgID int64, link *domain.Link) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback(ctx))
		}
	}()

	// Извлекаем url_id по url
	dsSelect := r.db.From("urls").
		Select("id").
		Where(goqu.Ex{"url": link.URL})

	sqlSelect, argsSelect, err := dsSelect.ToSQL()
	if err != nil {
		return err
	}

	var urlID int64
	if err := tx.QueryRow(ctx, sqlSelect, argsSelect...).Scan(&urlID); err != nil {
		return err
	}

	// Обновляем трек с новыми тегами и фильтрами
	dsUpdate := r.db.Update("tracks").
		Set(goqu.Record{
			"tags":    stringArrayToPostgres(link.Tags),
			"filters": stringArrayToPostgres(link.Filters),
		}).
		Where(goqu.Ex{"tg_id": tgID, "url_id": urlID})

	sqlUpdate, argsUpdate, err := dsUpdate.ToSQL()
	if err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, sqlUpdate, argsUpdate...); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// UpdateTimeLink обновляет время последнего обновления для url.
func (r *LinkRepoGoqu) UpdateTimeLink(ctx context.Context, lastUpdate time.Time, id int64) error {
	ds := r.db.Update("urls").
		Set(goqu.Record{"last_update": lastUpdate}).
		Where(goqu.Ex{"id": id})

	sql, args, err := ds.ToSQL()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, sql, args...)

	return err
}

func stringArrayToPostgres(arr []string) string {
	return "{" + strings.Join(arr, ",") + "}"
}
