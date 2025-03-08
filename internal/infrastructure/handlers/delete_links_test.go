package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	scrapperdto "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/dto/dto_scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/handlers"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/handlers/mocks"
)

func TestDeleteLinksHandler_InvalidChatID(t *testing.T) {
	mockScrapper := &mocks.Scrapper{}
	handler := handlers.DeleteLinksHandler{Scrapper: mockScrapper}

	req := httptest.NewRequest(http.MethodPost, "/tg-chat", http.NoBody)
	req.Header.Set("Tg-Chat-Id", "invalid_chat_id")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	body, _ := io.ReadAll(rr.Body)
	assert.Contains(t, string(body), "INVALID_CHAT_ID")
}

func TestDeleteLinksHandler_InvalidRequestBody(t *testing.T) {
	mockScrapper := &mocks.Scrapper{}
	handler := handlers.DeleteLinksHandler{Scrapper: mockScrapper}

	req := httptest.NewRequest(http.MethodPost, "/tg-chat", strings.NewReader("invalid json"))
	req.Header.Set("Tg-Chat-Id", "123")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	body, _ := io.ReadAll(rr.Body)
	assert.Contains(t, string(body), "INVALID_REQUEST_BODY")
}

func TestDeleteLinksHandler_UserNotExist(t *testing.T) {
	mockScrapper := &mocks.Scrapper{}
	tgChatID := int64(123)
	mockScrapper.
		On("DeleteLink", tgChatID, mock.MatchedBy(func(link domain.Link) bool {
			return link.URL == "https://github.com/user/repo"
		})).
		Return(domain.Link{}, domain.ErrUserNotExist{}).
		Once()

	handler := handlers.DeleteLinksHandler{Scrapper: mockScrapper}

	linkURL := "https://github.com/user/repo"
	removeReq := scrapperdto.RemoveLinkRequest{Link: &linkURL}
	reqBody, err := json.Marshal(removeReq)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/tg-chat", bytes.NewReader(reqBody))
	req.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgChatID, 10))

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	body, _ := io.ReadAll(rr.Body)
	assert.Contains(t, string(body), "CHAT_NOT_EXIST")

	mockScrapper.AssertExpectations(t)
}

func TestDeleteLinksHandler_DeleteLinkFailed(t *testing.T) {
	mockScrapper := &mocks.Scrapper{}
	tgChatID := int64(123)
	mockScrapper.
		On("DeleteLink", tgChatID, mock.MatchedBy(func(link domain.Link) bool {
			return link.URL == "https://github.com/user/repo"
		})).
		Return(domain.Link{}, errors.New("delete failed")).
		Once()

	handler := handlers.DeleteLinksHandler{Scrapper: mockScrapper}

	linkURL := "https://github.com/user/repo"
	removeReq := scrapperdto.RemoveLinkRequest{Link: &linkURL}
	reqBody, err := json.Marshal(removeReq)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/tg-chat", bytes.NewReader(reqBody))
	req.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgChatID, 10))

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	body, _ := io.ReadAll(rr.Body)
	assert.Contains(t, string(body), "DELETE_LINK_FAILED")

	mockScrapper.AssertExpectations(t)
}

func TestDeleteLinksHandler_Success(t *testing.T) {
	mockScrapper := &mocks.Scrapper{}
	tgChatID := int64(123)
	expectedLink := domain.Link{
		URL:     "https://github.com/user/repo",
		Tags:    []string{"tag1", "tag2"},
		Filters: []string{"filter1"},
		ID:      0,
	}

	mockScrapper.
		On("DeleteLink", tgChatID, mock.MatchedBy(func(link domain.Link) bool {
			return link.URL == expectedLink.URL
		})).
		Return(expectedLink, nil).
		Once()

	handler := handlers.DeleteLinksHandler{Scrapper: mockScrapper}

	linkURL := expectedLink.URL
	removeReq := scrapperdto.RemoveLinkRequest{Link: &linkURL}
	reqBody, err := json.Marshal(removeReq)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/tg-chat", bytes.NewReader(reqBody))
	req.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgChatID, 10))

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response scrapperdto.LinkResponse
	err = json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Url)
	assert.Equal(t, expectedLink.URL, *response.Url)
	assert.NotNil(t, response.Id)
	assert.Equal(t, tgChatID, *response.Id)
	assert.NotNil(t, response.Tags)
	assert.Equal(t, expectedLink.Tags, *response.Tags)
	assert.NotNil(t, response.Filters)
	assert.Equal(t, expectedLink.Filters, *response.Filters)

	mockScrapper.AssertExpectations(t)
}
