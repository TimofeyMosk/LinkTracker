package handlers_test

import (
	"errors"
	"fmt"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/handlers"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/handlers/mocks"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestDeleteUserHandler_InvalidChatID(t *testing.T) {
	mockScrapper := &mocks.Scrapper{}
	handler := handlers.DeleteUserHandler{Scrapper: mockScrapper}
	id := "errorID"
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s", id), nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	body, _ := io.ReadAll(rr.Body)
	assert.Contains(t, string(body), "INVALID_CHAT_ID")
}

func TestDeleteUserHandler_UserNotExist(t *testing.T) {
	mockScrapper := &mocks.Scrapper{}
	tgID := int64(123)
	mockScrapper.
		On("DeleteUser", tgID).
		Return(domain.ErrUserNotExist{}).
		Once()

	handler := handlers.DeleteUserHandler{Scrapper: mockScrapper}
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%d", tgID), nil)
	req.SetPathValue("id", strconv.Itoa(int(tgID)))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	body, _ := io.ReadAll(rr.Body)
	assert.Contains(t, string(body), "CHAT_NOT_EXIST")

	mockScrapper.AssertExpectations(t)
}

func TestDeleteUserHandler_ChatNotDeleted(t *testing.T) {
	mockScrapper := &mocks.Scrapper{}
	tgID := int64(123)
	genericErr := errors.New("deletion error")
	mockScrapper.
		On("DeleteUser", tgID).
		Return(genericErr).
		Once()

	handler := handlers.DeleteUserHandler{Scrapper: mockScrapper}
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%d", tgID), nil)
	req.SetPathValue("id", strconv.Itoa(int(tgID)))

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	body, _ := io.ReadAll(rr.Body)
	assert.Contains(t, string(body), "CHAT_NOT_DELETED")

	mockScrapper.AssertExpectations(t)
}

func TestDeleteUserHandler_Success(t *testing.T) {
	mockScrapper := &mocks.Scrapper{}
	tgID := int64(123)
	mockScrapper.
		On("DeleteUser", tgID).
		Return(nil).
		Once()

	handler := handlers.DeleteUserHandler{Scrapper: mockScrapper}
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%d", tgID), nil)
	req.SetPathValue("id", strconv.Itoa(int(tgID)))
	rr := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	mockScrapper.AssertExpectations(t)
}
