package goqurepo_test

import (
	"context"
	"testing"
	"time"

	"LinkTracker/internal/infrastructure/repository/postgresql"
	"LinkTracker/internal/infrastructure/repository/postgresql/goqurepo"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_UserRepo(t *testing.T) {
	// Настраиваем контекст и тестовую базу
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool, cleanup, err := postgresql.RunTestContainers(ctx)
	require.NoError(t, err)
	defer cleanup()

	userRepo := goqurepo.NewUserRepoGoqu(pool)

	t.Run("CreateUser and GetAllUsers", func(t *testing.T) {
		// Создаём тестового пользователя
		testUserID := int64(10001)
		// Перед созданием можно попытаться удалить пользователя, если он уже существует
		_ = userRepo.DeleteUser(ctx, testUserID)
		err := userRepo.CreateUser(ctx, testUserID)
		require.NoError(t, err, "Ошибка создания пользователя")

		// Получаем список всех пользователей
		users, err := userRepo.GetAllUsers(ctx)
		require.NoError(t, err, "Ошибка получения пользователей")
		// Проверяем, что наш пользователь присутствует в списке
		assert.Contains(t, users, testUserID, "Созданный пользователь не найден в выборке")
	})

	t.Run("DeleteUser", func(t *testing.T) {
		// Создаём пользователя, который будем удалять
		testUserID := int64(10002)
		err := userRepo.CreateUser(ctx, testUserID)
		require.NoError(t, err, "Ошибка создания пользователя для удаления")

		// Удаляем пользователя
		err = userRepo.DeleteUser(ctx, testUserID)
		require.NoError(t, err, "Ошибка удаления пользователя")

		// Получаем всех пользователей и убеждаемся, что удалённого пользователя в списке нет
		users, err := userRepo.GetAllUsers(ctx)
		require.NoError(t, err, "Ошибка получения пользователей после удаления")
		assert.NotContains(t, users, testUserID, "Пользователь не был удалён")
	})

	t.Run("DeleteNonExistentUser", func(t *testing.T) {
		// Пытаемся удалить несуществующего пользователя
		nonExistentUserID := int64(99999)
		err := userRepo.DeleteUser(ctx, nonExistentUserID)
		assert.Equal(t, pgx.ErrNoRows, err, "Ожидаем ошибку pgx.ErrNoRows для несуществующего пользователя")
	})
}
