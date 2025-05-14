package scrapper_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"LinkTracker/internal/application/scrapper"
	"LinkTracker/internal/application/scrapper/mocks"
	"LinkTracker/internal/domain"
)

func Test_Scrapper_AddUser_Success(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)

	userRepo.On("CreateUser", ctx, tgID).Return(nil)
	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	err := s.AddUser(ctx, tgID)

	assert.Nil(t, err)
	userRepo.AssertExpectations(t)
}

func Test_Scrapper_AddUser_Error(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)

	userRepo.On("CreateUser", ctx, tgID).Return(errors.New("some error"))
	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	err := s.AddUser(ctx, tgID)

	assert.Error(t, err)
	userRepo.AssertExpectations(t)
}
func Test_Scrapper_DeleteUser_Success(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)

	userRepo.On("DeleteUser", ctx, tgID).Return(nil)
	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	err := s.DeleteUser(ctx, tgID)

	assert.Nil(t, err)
	userRepo.AssertExpectations(t)
}

func Test_Scrapper_DeleteUser_Error(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)

	userRepo.On("DeleteUser", ctx, tgID).Return(errors.New("some error"))
	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	err := s.DeleteUser(ctx, tgID)

	assert.Error(t, err)
	userRepo.AssertExpectations(t)
}

func Test_Scrapper_GetLinks_Success(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)
	links := []domain.Link{{URL: "https://example/example", ID: 1}, {URL: "https://example/example2", ID: 2}}

	linkRepo.On("GetUserLinks", ctx, tgID).Return(links, nil)
	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	links, err := s.GetUserLinks(ctx, tgID)

	assert.Nil(t, err)
	assert.Equal(t, links, links)
	linkRepo.AssertExpectations(t)
}

func Test_Scrapper_GetLinks_Error(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)

	linkRepo.On("GetUserLinks", ctx, tgID).Return(nil, errors.New("some error"))
	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	links, err := s.GetUserLinks(ctx, tgID)

	assert.Error(t, err)
	assert.Nil(t, links)
	linkRepo.AssertExpectations(t)
}

func Test_Scrapper_AddLink_Success(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)
	newLink := domain.Link{URL: "https://example/example"}
	newLinkWithID := domain.Link{URL: "https://example/example", ID: 1}

	linkRepo.On("GetUserLinks", ctx, tgID).Return([]domain.Link{}, nil)
	linkRepo.On("AddLink", ctx, tgID, &newLink).Return(newLinkWithID, nil)

	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	link, err := s.AddLink(ctx, tgID, &newLink)
	assert.Nil(t, err)
	assert.Equal(t, newLinkWithID, link)
	linkRepo.AssertExpectations(t)
}

func Test_Scrapper_AddLink_GetUserLinksError(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)
	newLink := domain.Link{URL: "https://example/example"}

	linkRepo.On("GetUserLinks", ctx, tgID).Return(nil, errors.New("some error"))

	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	link, err := s.AddLink(ctx, tgID, &newLink)
	assert.Error(t, err)
	assert.Equal(t, domain.Link{}, link)
	linkRepo.AssertExpectations(t)
}

func Test_Scrapper_AddLink_LinkAlreadyTracks(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)
	newLink := domain.Link{URL: "https://example/example"}

	linkRepo.On("GetUserLinks", ctx, tgID).Return([]domain.Link{newLink}, nil)

	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	link, err := s.AddLink(ctx, tgID, &newLink)
	assert.ErrorIs(t, err, domain.ErrLinkAlreadyTracking{})
	assert.Equal(t, domain.Link{}, link)
	linkRepo.AssertExpectations(t)
}

func Test_Scrapper_AddLink_Error(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)
	newLink := domain.Link{URL: "https://example/example"}

	linkRepo.On("GetUserLinks", ctx, tgID).Return([]domain.Link{}, nil)
	linkRepo.On("AddLink", ctx, tgID, &newLink).Return(domain.Link{}, errors.New("some error"))

	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	link, err := s.AddLink(ctx, tgID, &newLink)
	assert.Error(t, err)
	assert.Equal(t, domain.Link{}, link)
	linkRepo.AssertExpectations(t)
}

func Test_Scrapper_DeleteLink_Success(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)
	link := domain.Link{URL: "https://example/example"}

	linkRepo.On("DeleteLink", ctx, tgID, &link).Return(link, nil)

	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	deletedLink, err := s.DeleteLink(ctx, tgID, &link)
	assert.Nil(t, err)
	assert.Equal(t, link, deletedLink)
	linkRepo.AssertExpectations(t)
}
func Test_Scrapper_DeleteLink_LinkNotExistError(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)
	link := domain.Link{URL: "https://example/example"}

	linkRepo.On("DeleteLink", ctx, tgID, &link).Return(domain.Link{}, domain.ErrLinkNotExist{})

	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	deletedLink, err := s.DeleteLink(ctx, tgID, &link)
	assert.ErrorIs(t, err, domain.ErrLinkNotExist{})
	assert.Equal(t, domain.Link{}, deletedLink)
	linkRepo.AssertExpectations(t)
}

func Test_Scrapper_DeleteLink_Error(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)
	link := domain.Link{URL: "https://example/example"}

	linkRepo.On("DeleteLink", ctx, tgID, &link).Return(domain.Link{}, errors.New("some error"))

	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	deletedLink, err := s.DeleteLink(ctx, tgID, &link)
	assert.Error(t, err)
	assert.Equal(t, domain.Link{}, deletedLink)
	linkRepo.AssertExpectations(t)
}

func Test_Scrapper_UpdateLink_Success(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)
	link := domain.Link{URL: "https://example/example"}

	linkRepo.On("UpdateLink", ctx, tgID, &link).Return(nil)

	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	err := s.UpdateLink(ctx, tgID, &link)
	assert.Nil(t, err)
	linkRepo.AssertExpectations(t)
}

func Test_Scrapper_UpdateLink_Error(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)
	link := domain.Link{URL: "https://example/example"}

	linkRepo.On("UpdateLink", ctx, tgID, &link).Return(errors.New("some error"))

	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	err := s.UpdateLink(ctx, tgID, &link)
	assert.Error(t, err)
	linkRepo.AssertExpectations(t)
}

func Test_Scrapper_CreateState_Success(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)
	state := 1

	stateRepo.On("CreateState", ctx, tgID, state).Return(nil)

	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	err := s.CreateState(ctx, tgID, state)
	assert.Nil(t, err)
	linkRepo.AssertExpectations(t)
}

func Test_Scrapper_CreateState_Error(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)
	state := 1

	stateRepo.On("CreateState", ctx, tgID, state).Return(errors.New("some error"))

	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	err := s.CreateState(ctx, tgID, state)
	assert.Error(t, err)
	linkRepo.AssertExpectations(t)
}

func Test_Scrapper_DeleteState_Success(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)

	stateRepo.On("DeleteState", ctx, tgID).Return(nil)

	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	err := s.DeleteState(ctx, tgID)
	assert.Nil(t, err)
	linkRepo.AssertExpectations(t)
}

func Test_Scrapper_DeleteState_Error(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)

	stateRepo.On("DeleteState", ctx, tgID).Return(errors.New("some error"))

	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	err := s.DeleteState(ctx, tgID)
	assert.Error(t, err)
	linkRepo.AssertExpectations(t)
}

func Test_Scrapper_GetState_Success(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)
	state := 1
	link := domain.Link{URL: "https://example/example"}

	stateRepo.On("GetState", ctx, tgID).Return(state, link, nil)

	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	getState, getLink, err := s.GetState(ctx, tgID)
	assert.Nil(t, err)
	assert.Equal(t, state, getState)
	assert.Equal(t, link, getLink)
	linkRepo.AssertExpectations(t)
}

func Test_Scrapper_GetState_Error(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)

	stateRepo.On("GetState", ctx, tgID).Return(-1, domain.Link{}, errors.New("some error"))

	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	getState, getLink, err := s.GetState(ctx, tgID)
	assert.Error(t, err)
	assert.Equal(t, -1, getState)
	assert.Equal(t, domain.Link{}, getLink)
	linkRepo.AssertExpectations(t)
}

func Test_Scrapper_UpdateState_Success(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)
	state := 1
	link := domain.Link{URL: "https://example/example"}

	stateRepo.On("UpdateState", ctx, tgID, state, &link).Return(nil)

	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	err := s.UpdateState(ctx, tgID, state, &link)
	assert.Nil(t, err)

	linkRepo.AssertExpectations(t)
}
func Test_Scrapper_UpdateState_Error(t *testing.T) {
	ctx := context.Background()
	interval := 1 * time.Minute
	linkRepo := &mocks.LinkRepo{}
	userRepo := &mocks.UserRepo{}
	stateRepo := &mocks.StateRepo{}
	notifier := &mocks.Notifier{}
	linkChecker := &mocks.LinkChecker{}
	tgID := int64(123)
	state := 1
	link := domain.Link{URL: "https://example/example"}

	stateRepo.On("UpdateState", ctx, tgID, state, &link).Return(errors.New("some error"))

	s := scrapper.NewScrapper(userRepo, linkRepo, stateRepo, interval, notifier, linkChecker)

	err := s.UpdateState(ctx, tgID, state, &link)
	assert.Error(t, err)

	linkRepo.AssertExpectations(t)
}
