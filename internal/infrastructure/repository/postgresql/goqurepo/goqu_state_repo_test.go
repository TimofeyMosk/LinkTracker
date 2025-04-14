package goqurepo_test

import (
	"context"
	"testing"
	"time"

	"LinkTracker/internal/infrastructure/repository/postgresql"
	"LinkTracker/internal/infrastructure/repository/postgresql/goqurepo"

	"LinkTracker/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_StateRepo(t *testing.T) {
	// Настраиваем контекст и тестовую базу данных.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool, cleanup, err := postgresql.RunPostgresAndMigrateTestContainers(ctx)
	require.NoError(t, err)
	defer cleanup()

	stateRepo := goqurepo.NewStateRepoGoqu(pool)

	// Используем некоторое тестовое значение tg_id
	const tgID int64 = 55555

	t.Run("CreateState and GetState", func(t *testing.T) {
		// Создаём состояние для пользователя (без дополнительных данных для Link)
		initialState := 1
		err := stateRepo.CreateState(ctx, tgID, initialState)
		require.NoError(t, err)

		// Получаем состояние и проверяем, что оно соответствует ожидаемому.
		// Поскольку при создании через CreateState передаётся только state, остальные поля (url, tags, filters)
		// должны иметь значения по умолчанию (пустая строка и пустые срезы)
		state, link, err := stateRepo.GetState(ctx, tgID)
		require.NoError(t, err)
		assert.Equal(t, initialState, state)
		assert.Equal(t, "", link.URL, "Ожидается, что URL по умолчанию пустой")
		assert.Empty(t, link.Tags, "Ожидается, что Tags по умолчанию пустой срез")
		assert.Empty(t, link.Filters, "Ожидается, что Filters по умолчанию пустой срез")
	})

	t.Run("UpdateState", func(t *testing.T) {
		// Обновляем состояние и связанные данные.
		updatedState := 2
		updateLink := &domain.Link{
			URL:     "http://state.example.com",
			Tags:    []string{"state-tag1", "state-tag2"},
			Filters: []string{"state-filter1"},
		}
		err := stateRepo.UpdateState(ctx, tgID, updatedState, updateLink)
		require.NoError(t, err)

		// Проверяем, что обновление прошло успешно.
		state, link, err := stateRepo.GetState(ctx, tgID)
		require.NoError(t, err)
		assert.Equal(t, updatedState, state, "Ожидается обновлённое состояние")
		assert.Equal(t, updateLink.URL, link.URL, "URL не обновился")
		assert.Equal(t, updateLink.Tags, link.Tags, "Tags не обновились")
		assert.Equal(t, updateLink.Filters, link.Filters, "Filters не обновились")
	})

	t.Run("DeleteState", func(t *testing.T) {
		// Удаляем состояние для данного tg_id
		err := stateRepo.DeleteState(ctx, tgID)
		require.NoError(t, err)

		// После удаления попытка получить состояние должна вернуть ошибку, поскольку запись отсутствует.
		_, _, err = stateRepo.GetState(ctx, tgID)
		require.Error(t, err, "Ожидается ошибка при получении несуществующего состояния")
		// Дополнительно можно проверить, что ошибка соответствует pgx.ErrNoRows
		assert.Equal(t, pgx.ErrNoRows, err, "Ожидается pgx.ErrNoRows при отсутствии записи")
	})
}
