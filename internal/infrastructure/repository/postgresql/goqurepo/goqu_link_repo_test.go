package goqurepo_test

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"LinkTracker/internal/infrastructure/repository/postgresql/goqurepo"

	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"LinkTracker/internal/domain"
	"LinkTracker/internal/infrastructure/repository/postgresql"
)

// helperInsertUser проверяет, что пользователь с данным tg_id существует, и если нет, вставляет его.
func helperInsertUser(ctx context.Context, t *testing.T, pool *pgxpool.Pool, tgID int64) {
	// Используем ON CONFLICT DO NOTHING, чтобы избежать ошибки при повторном запуске
	_, err := pool.Exec(ctx, `INSERT INTO users (tg_id) VALUES ($1) ON CONFLICT DO NOTHING`, tgID)
	require.NoError(t, err, "Не удалось вставить/подтвердить пользователя")
}

func Test_LinkRepo(t *testing.T) {
	// Настраиваем контекст и тестовую базу
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool, cleanup, err := postgresql.RunTestContainers(ctx)
	defer cleanup()
	require.NoError(t, err)

	linkRepo := goqurepo.NewLinkRepoGoqu(pool)

	// Идентификатор тестового пользователя
	const tgID int64 = 12345

	// Вставляем пользователя, так как таблица tracks ссылается на users
	helperInsertUser(ctx, t, pool, tgID)

	// Тестовая ссылка
	var testLink domain.Link
	testLink.URL = "http://example.com"
	testLink.Filters = []string{"filter1"}
	testLink.Tags = []string{"tag1"}

	t.Run("Add and Get User Links", func(t *testing.T) {
		// Добавляем ссылку для пользователя
		err := linkRepo.AddLink(ctx, tgID, &testLink)
		require.NoError(t, err)

		// Получаем ссылки пользователя и проверяем, что добавленная ссылка присутствует
		links, err := linkRepo.GetUserLinks(ctx, tgID)
		require.NoError(t, err)
		require.NotEmpty(t, links)

		var found bool

		for _, l := range links {
			if l.URL == testLink.URL {
				found = true
				// Сохраняем идентификатор для последующих тестов
				testLink.ID = l.ID

				break
			}
		}

		assert.True(t, found, "Добавленная ссылка не найдена у пользователя")
	})

	t.Run("Get Users By Link", func(t *testing.T) {
		// Проверяем, что по ID ссылки возвращается нужный пользователь
		require.NotZero(t, testLink.ID, "ID ссылки должен быть установлен")
		tgIDs, err := linkRepo.GetUsersByLink(ctx, testLink.ID)
		require.NoError(t, err)
		assert.Contains(t, tgIDs, tgID, "Пользователь не найден по ссылке")
	})

	t.Run("Update Link", func(t *testing.T) {
		// Обновляем фильтры и теги для ранее добавленной ссылки
		updatedFilters := []string{"updated_filter"}
		updatedTags := []string{"updated_tag"}

		updateLink := &domain.Link{
			URL:     testLink.URL,
			Filters: updatedFilters,
			Tags:    updatedTags,
		}
		err := linkRepo.UpdateLink(ctx, tgID, updateLink)
		require.NoError(t, err)

		// Проверяем, что обновление прошло успешно
		links, err := linkRepo.GetUserLinks(ctx, tgID)
		require.NoError(t, err)

		var found *domain.Link

		for _, l := range links {
			if l.URL == testLink.URL {
				found = &l
				break
			}
		}

		require.NotNil(t, found, "Ссылка не найдена после обновления")
		assert.Equal(t, updatedFilters, found.Filters, "Фильтры не обновились")
		assert.Equal(t, updatedTags, found.Tags, "Теги не обновились")
	})

	t.Run("Update Time Link", func(t *testing.T) {
		// Обновляем время последнего обновления для ссылки
		newTime := time.Now().UTC()
		err := linkRepo.UpdateTimeLink(ctx, newTime, testLink.ID)
		require.NoError(t, err)

		// Получаем все ссылки для проверки времени обновления
		allLinks, err := linkRepo.GetAllLinks(ctx)
		require.NoError(t, err)

		var foundLink *domain.Link

		for _, l := range allLinks {
			if l.ID == testLink.ID {
				foundLink = &l
				break
			}
		}

		require.NotNil(t, foundLink, "Ссылка не найдена после обновления времени")
		assert.WithinDuration(t, newTime, foundLink.LastUpdated, time.Second, "Время обновления не совпадает")
	})

	t.Run("Get Links After", func(t *testing.T) {
		// Выбираем ссылки, обновленные после определённого времени (например, за последний час)
		pastTime := time.Now().UTC().Add(-time.Hour)
		linksAfter, err := linkRepo.GetLinksAfter(ctx, pastTime, 10)
		require.NoError(t, err)
		require.NotEmpty(t, linksAfter, "Ожидается, что будут ссылки после заданного времени")

		var found bool

		for _, l := range linksAfter {
			if l.ID == testLink.ID {
				found = true
				break
			}
		}

		assert.True(t, found, "Ожидаемая ссылка не найдена в выборке по времени")
	})

	t.Run("Delete Link", func(t *testing.T) {
		// Удаляем ссылку для пользователя
		deletedLink, err := linkRepo.DeleteLink(ctx, tgID, &domain.Link{URL: testLink.URL})
		require.NoError(t, err)
		assert.Equal(t, testLink.URL, deletedLink.URL, "URL удалённой ссылки не совпадает")

		// Проверяем, что ссылка отсутствует в списке треков пользователя
		links, err := linkRepo.GetUserLinks(ctx, tgID)
		require.NoError(t, err)

		for _, l := range links {
			assert.NotEqual(t, testLink.URL, l.URL, "Ссылка не удалена из списка треков пользователя")
		}

		// Проверяем, что ссылка удалена из общей таблицы urls, если треков для неё не осталось
		allLinks, err := linkRepo.GetAllLinks(ctx)
		require.NoError(t, err)

		for _, l := range allLinks {
			assert.NotEqual(t, testLink.URL, l.URL, "Ссылка не удалена из таблицы urls")
		}
	})
}
