package handlers_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	scrapperdto "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/dto/dto_scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/handlers"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/handlers/mocks"
)

func TestGetLinksHandler_InvalidChatID(t *testing.T) {
	mockScrapper := &mocks.Scrapper{}
	handler := handlers.GetLinksHandler{Scrapper: mockScrapper}

	req := httptest.NewRequest(http.MethodGet, "/links", http.NoBody)
	req.Header.Set("Tg-Chat-Id", "invalid-chat-id")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	body, err := io.ReadAll(rr.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "INVALID_CHAT_ID")
}

func TestGetLinksHandler_UserNotExist(t *testing.T) {
	mockScrapper := &mocks.Scrapper{}
	tgID := int64(123)
	mockScrapper.
		On("GetLinks", tgID).
		Return(nil, domain.ErrUserNotExist{}).
		Once()

	handler := handlers.GetLinksHandler{Scrapper: mockScrapper}
	req := httptest.NewRequest(http.MethodGet, "/links", http.NoBody)
	req.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	body, err := io.ReadAll(rr.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "CHAT_NOT_EXIST")

	mockScrapper.AssertExpectations(t)
}

func TestGetLinksHandler_GenericError(t *testing.T) {
	mockScrapper := &mocks.Scrapper{}
	tgID := int64(123)
	genErr := errors.New("some error")
	mockScrapper.
		On("GetLinks", tgID).
		Return(nil, genErr).
		Once()

	handler := handlers.GetLinksHandler{Scrapper: mockScrapper}
	req := httptest.NewRequest(http.MethodGet, "/links", http.NoBody)
	req.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	body, err := io.ReadAll(rr.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "LINKS_NOT_RECEIVED")

	mockScrapper.AssertExpectations(t)
}

func TestGetLinksHandler_Success(t *testing.T) {
	mockScrapper := &mocks.Scrapper{}
	tgID := int64(123)
	expectedLinks := []domain.Link{
		{
			URL:     "https://github.com/test/repo",
			Tags:    []string{"go", "api"},
			Filters: []string{"active"},
			ID:      1,
		},
	}
	mockScrapper.
		On("GetLinks", tgID).
		Return(expectedLinks, nil).
		Once()

	handler := handlers.GetLinksHandler{Scrapper: mockScrapper}
	req := httptest.NewRequest(http.MethodGet, "/links", http.NoBody)
	req.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var responseData scrapperdto.ListLinksResponse
	err := json.NewDecoder(rr.Body).Decode(&responseData)
	require.NoError(t, err)
	require.Equal(t, *responseData.Size, int32(len(expectedLinks))) //nolint:gosec // overflow impossibly
	assert.Equal(t, expectedLinks[0].URL, *(*responseData.Links)[0].Url)

	mockScrapper.AssertExpectations(t)
}
