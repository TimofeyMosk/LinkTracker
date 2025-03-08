package handlers_test

import (
	"errors"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/handlers"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/handlers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostUserHandler_Success(t *testing.T) {
	mockScrapper := &mocks.Scrapper{}
	mockScrapper.On("AddUser", mock.Anything).Return(nil)

	handler := handlers.PostUserHandler{Scrapper: mockScrapper}

	req := httptest.NewRequest(http.MethodPost, "/users/123", nil)
	req.SetPathValue("id", "123")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Body.String())

	mockScrapper.AssertExpectations(t)
}

func TestPostUserHandler_InvalidChatID(t *testing.T) {
	mockScrapper := &mocks.Scrapper{}

	handler := handlers.PostUserHandler{Scrapper: mockScrapper}

	req := httptest.NewRequest(http.MethodPost, "/users/invalid", nil)
	req.SetPathValue("id", "invalid")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

}

func TestPostUserHandler_ScrapperError(t *testing.T) {
	mockScrapper := &mocks.Scrapper{}
	mockScrapper.On("AddUser", mock.Anything).Return(errors.New("some error"))

	handler := handlers.PostUserHandler{Scrapper: mockScrapper}

	req := httptest.NewRequest(http.MethodPost, "/users/123", nil)
	req.SetPathValue("id", "123")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	mockScrapper.AssertExpectations(t)
}
