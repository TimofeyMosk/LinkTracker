package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	scrapperdto "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/dto/dto_scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/handlers"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/handlers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestPostLinkHandler_InvalidChatID(t *testing.T) {
	mockScrapper := &mocks.Scrapper{}
	handler := handlers.PostLinkHandler{Scrapper: mockScrapper}
	req := httptest.NewRequest(http.MethodPost, "/links", nil)
	req.Header.Set("Tg-Chat-Id", "invalid")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	body, err := io.ReadAll(rr.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "INVALID_CHAT_ID")
}

func TestPostLinkHandler_InvalidRequestBody(t *testing.T) {
	mockScrapper := &mocks.Scrapper{}
	handler := handlers.PostLinkHandler{Scrapper: mockScrapper}
	req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Tg-Chat-Id", "123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	body, err := io.ReadAll(rr.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "INVALID_REQUEST_BODY")
}

func TestPostLinkHandler_InvalidLinkConversion(t *testing.T) {
	reqBody, err := json.Marshal(scrapperdto.AddLinkRequest{
		Link:    nil,
		Tags:    &[]string{"tag1", "tag2"},
		Filters: &[]string{"filter1"},
	})
	require.NoError(t, err)

	mockScrapper := &mocks.Scrapper{}
	handler := handlers.PostLinkHandler{Scrapper: mockScrapper}
	req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewReader(reqBody))
	req.Header.Set("Tg-Chat-Id", "123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	body, err := io.ReadAll(rr.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "INVALID_REQUEST_BODY")
}

func TestPostLinkHandler_ScrapperError_ChatNotExist(t *testing.T) {
	tgChatID := int64(123)
	linkURL := "https://github.com/user/repo"
	reqData := scrapperdto.AddLinkRequest{
		Link:    &linkURL,
		Tags:    &[]string{"tag1", "tag2"},
		Filters: &[]string{"filter1"},
	}
	reqBody, err := json.Marshal(reqData)
	require.NoError(t, err)

	mockScrapper := &mocks.Scrapper{}
	mockScrapper.
		On("AddLink", tgChatID, mock.MatchedBy(func(l domain.Link) bool {
			return l.URL == linkURL
		})).
		Return(domain.Link{}, domain.ErrUserNotExist{}).
		Once()

	handler := handlers.PostLinkHandler{Scrapper: mockScrapper}
	req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewReader(reqBody))
	req.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgChatID, 10))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	body, err := io.ReadAll(rr.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "CHAT_NOT_EXIST")
	mockScrapper.AssertExpectations(t)
}

func TestPostLinkHandler_ScrapperError_Generic(t *testing.T) {
	tgChatID := int64(123)
	linkURL := "https://github.com/user/repo"
	reqData := scrapperdto.AddLinkRequest{
		Link:    &linkURL,
		Tags:    &[]string{"tag1", "tag2"},
		Filters: &[]string{"filter1"},
	}
	reqBody, err := json.Marshal(reqData)
	require.NoError(t, err)

	mockScrapper := &mocks.Scrapper{}
	genErr := errors.New("some error")
	mockScrapper.
		On("AddLink", tgChatID, mock.MatchedBy(func(l domain.Link) bool {
			return l.URL == linkURL
		})).
		Return(domain.Link{}, genErr).
		Once()

	handler := handlers.PostLinkHandler{Scrapper: mockScrapper}
	req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewReader(reqBody))
	req.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgChatID, 10))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	body, err := io.ReadAll(rr.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "ADD_LINK_FAILED")
	mockScrapper.AssertExpectations(t)
}

func TestPostLinkHandler_Success(t *testing.T) {
	tgChatID := int64(123)
	linkURL := "https://github.com/user/repo"
	reqData := scrapperdto.AddLinkRequest{
		Link:    &linkURL,
		Tags:    &[]string{"tag1", "tag2"},
		Filters: &[]string{"filter1"},
	}
	reqBody, err := json.Marshal(reqData)
	require.NoError(t, err)

	expectedLink := domain.Link{
		ID:      1,
		URL:     linkURL,
		Tags:    []string{"tag1", "tag2"},
		Filters: []string{"filter1"},
	}
	mockScrapper := &mocks.Scrapper{}
	mockScrapper.
		On("AddLink", tgChatID, mock.MatchedBy(func(l domain.Link) bool {
			return l.URL == linkURL
		})).
		Return(expectedLink, nil).
		Once()

	handler := handlers.PostLinkHandler{Scrapper: mockScrapper}
	req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewReader(reqBody))
	req.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgChatID, 10))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	var response scrapperdto.LinkResponse
	err = json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.NotNil(t, response.Id)
	assert.Equal(t, expectedLink.ID, *response.Id)
	assert.NotNil(t, response.Url)
	assert.Equal(t, expectedLink.URL, *response.Url)
	assert.NotNil(t, response.Tags)
	assert.Equal(t, expectedLink.Tags, *response.Tags)
	assert.NotNil(t, response.Filters)
	assert.Equal(t, expectedLink.Filters, *response.Filters)
	mockScrapper.AssertExpectations(t)
}
