package repository_test

import (
	"testing"

	"LinkTracker/internal/domain"
	"LinkTracker/internal/infrastructure/repository"

	"github.com/stretchr/testify/assert"
)

func Test_Repository_NewRepository(t *testing.T) {
	repo := repository.NewRepository()

	assert.NotNil(t, repo)
	assert.Len(t, repo.Links, 0)
}

func Test_Repository_CreateUser(t *testing.T) {
	repo := repository.NewRepository()
	userID := int64(1)

	err := repo.CreateUser(userID)

	assert.NoError(t, err)
	assert.True(t, repo.UserExist(userID))
}

func Test_Repository_CreateUser_AlreadyExists(t *testing.T) {
	repo := repository.NewRepository()
	userID := int64(1)
	_ = repo.CreateUser(userID)

	err := repo.CreateUser(userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func Test_Repository_DeleteUser(t *testing.T) {
	repo := repository.NewRepository()
	userID := int64(1)
	_ = repo.CreateUser(userID)

	err := repo.DeleteUser(userID)

	assert.NoError(t, err)
	assert.False(t, repo.UserExist(userID))
}

func Test_Repository_DeleteUser_NotFound(t *testing.T) {
	repo := repository.NewRepository()
	userID := int64(1)

	err := repo.DeleteUser(userID)

	assert.Error(t, err)
	assert.ErrorAs(t, err, &domain.ErrUserNotExist{})
}

func Test_Repository_GetAllUsers(t *testing.T) {
	repo := repository.NewRepository()
	userID1 := int64(1)
	userID2 := int64(2)
	_ = repo.CreateUser(userID1)
	_ = repo.CreateUser(userID2)

	users, err := repo.GetAllUsers()

	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Contains(t, users, userID1)
	assert.Contains(t, users, userID2)
}

func Test_Repository_GetLinks(t *testing.T) {
	repo := repository.NewRepository()
	userID := int64(1)
	link1 := domain.Link{URL: "https://example.com/1"}
	link2 := domain.Link{URL: "https://example.com/2"}
	_ = repo.CreateUser(userID)
	_ = repo.AddLink(userID, link1)
	_ = repo.AddLink(userID, link2)

	links, err := repo.GetUserLinks(userID)

	assert.NoError(t, err)
	assert.Len(t, links, 2)
	assert.Contains(t, links, link1)
	assert.Contains(t, links, link2)
}

func Test_Repository_GetLinks_UserNotFound(t *testing.T) {
	repo := repository.NewRepository()
	userID := int64(1)

	_, err := repo.GetUserLinks(userID)

	assert.Error(t, err)
	assert.ErrorAs(t, err, &domain.ErrUserNotExist{})
}

func Test_Repository_AddLink(t *testing.T) {
	repo := repository.NewRepository()
	userID := int64(1)
	link := domain.Link{URL: "https://example.com"}
	_ = repo.CreateUser(userID)

	err := repo.AddLink(userID, link)

	assert.NoError(t, err)

	links, _ := repo.GetUserLinks(userID)
	assert.Len(t, links, 1)
	assert.Equal(t, links[0], link)
}

func Test_Repository_AddLink_UserNotFound(t *testing.T) {
	repo := repository.NewRepository()
	userID := int64(1)
	link := domain.Link{URL: "https://example.com/1"}

	err := repo.AddLink(userID, link)

	assert.Error(t, err)
	assert.ErrorAs(t, err, &domain.ErrUserNotExist{})
}

func Test_Repository_DeleteLink(t *testing.T) {
	repo := repository.NewRepository()
	userID := int64(1)
	link := domain.Link{URL: "https://example.com/1"}
	_ = repo.CreateUser(userID)
	_ = repo.AddLink(userID, link)

	deletedLink, err := repo.DeleteLink(userID, link)

	assert.NoError(t, err)
	assert.Equal(t, deletedLink, link)

	links, _ := repo.GetUserLinks(userID)
	assert.Len(t, links, 0)
}

func Test_Repository_DeleteLink_UserNotFound(t *testing.T) {
	repo := repository.NewRepository()
	userID := int64(1)
	link := domain.Link{URL: "https://example.com/1"}

	_, err := repo.DeleteLink(userID, link)

	assert.Error(t, err)
	assert.ErrorAs(t, err, &domain.ErrUserNotExist{})
}

func Test_Repository_DeleteLink_LinkNotFound(t *testing.T) {
	repo := repository.NewRepository()
	userID := int64(1)
	link := domain.Link{URL: "https://example.com/1"}

	_ = repo.CreateUser(userID)

	_, err := repo.DeleteLink(userID, link)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func Test_Repository_GetAllLinks(t *testing.T) {
	repo := repository.NewRepository()
	userID1 := int64(1)
	userID2 := int64(2)
	link1 := domain.Link{URL: "https://example.com/1"}
	link2 := domain.Link{URL: "https://example.com/2"}
	_ = repo.CreateUser(userID1)
	_ = repo.CreateUser(userID2)
	_ = repo.AddLink(userID1, link1)
	_ = repo.AddLink(userID2, link2)

	links, err := repo.GetAllLinks()

	assert.NoError(t, err)
	assert.Len(t, links, 2)
	assert.Contains(t, links, link1)
	assert.Contains(t, links, link2)
}

func Test_Repository_GetAllLinks_Empty(t *testing.T) {
	repo := repository.NewRepository()

	links, err := repo.GetAllLinks()

	assert.NoError(t, err)
	assert.Len(t, links, 0)
}
