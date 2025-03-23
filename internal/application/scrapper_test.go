package application_test

// import (
//	"testing"
//	"time"
//
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//
//	"github.com/es-debug/backend-academy-2024-go-template/internal/application"
//	"github.com/es-debug/backend-academy-2024-go-template/internal/application/mocks"
//	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
//)
//
// func TestScrapper_AddUser_Success(t *testing.T) {
//	dbMock := &mocks.Database{}
//	botClientMock := &mocks.BotClient{}
//
//	dbMock.On("CreateUser", int64(123)).Return(nil)
//
//	s := application.NewScrapper(dbMock, 1*time.Second, botClientMock)
//
//	err := s.AddUser(123)
//
//	require.NoError(t, err)
//	dbMock.AssertExpectations(t)
//}
//
// func TestScrapper_DeleteUser_Success(t *testing.T) {
//	dbMock := &mocks.Database{}
//	botClientMock := &mocks.BotClient{}
//
//	dbMock.On("DeleteUser", int64(123)).Return(nil)
//
//	s := application.NewScrapper(dbMock, 1*time.Second, botClientMock)
//
//	err := s.DeleteUser(123)
//
//	require.NoError(t, err)
//	dbMock.AssertExpectations(t)
//}
//
// func TestScrapper_GetLinks_Success(t *testing.T) {
//	dbMock := &mocks.Database{}
//	botClientMock := &mocks.BotClient{}
//	expectedLinks := []domain.Link{
//		{ID: 1, URL: "https://github.com/owner/repo", Tags: []string{"go"}, Filters: []string{"active"}},
//	}
//	dbMock.On("GetUserLinks", int64(123)).Return(expectedLinks, nil)
//
//	s := application.NewScrapper(dbMock, 1*time.Second, botClientMock)
//
//	links, err := s.GetUserLinks(123)
//
//	require.NoError(t, err)
//	assert.Equal(t, expectedLinks, links)
//	dbMock.AssertExpectations(t)
//}
//
// func TestScrapper_AddLink_Success(t *testing.T) {
//	dbMock := &mocks.Database{}
//	botClientMock := &mocks.BotClient{}
//	link := domain.Link{ID: 1, URL: "https://github.com/owner/repo", Tags: []string{"go"}, Filters: []string{"active"}}
//	dbMock.On("AddLink", int64(123), link).Return(nil)
//
//	s := application.NewScrapper(dbMock, 1*time.Second, botClientMock)
//
//	returnedLink, err := s.AddLink(123, link)
//
//	require.NoError(t, err)
//	assert.Equal(t, link, returnedLink)
//	dbMock.AssertExpectations(t)
//}
//
// func TestScrapper_AddLink_Invalid(t *testing.T) {
//	dbMock := &mocks.Database{}
//	botClientMock := &mocks.BotClient{}
//	invalidLink := domain.Link{ID: 1, URL: "https://google.com", Tags: []string{}, Filters: []string{}}
//
//	s := application.NewScrapper(dbMock, 1*time.Second, botClientMock)
//
//	_, err := s.AddLink(123, invalidLink)
//
//	require.Error(t, err)
//	assert.Equal(t, domain.ErrWrongURL{}, err)
//}
//
// func TestScrapper_DeleteLink_Success(t *testing.T) {
//	dbMock := &mocks.Database{}
//	botClientMock := &mocks.BotClient{}
//	link := domain.Link{ID: 1, URL: "https://github.com/owner/repo", Tags: []string{"go"}, Filters: []string{"active"}}
//	dbMock.On("DeleteLink", int64(123), link).Return(link, nil)
//
//	s := application.NewScrapper(dbMock, 1*time.Second, botClientMock)
//
//	returnedLink, err := s.DeleteLink(123, link)
//
//	require.NoError(t, err)
//	assert.Equal(t, link, returnedLink)
//	dbMock.AssertExpectations(t)
//}
