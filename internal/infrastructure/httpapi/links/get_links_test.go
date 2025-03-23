package links_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"LinkTracker/internal/domain"
	scrapperdto "LinkTracker/internal/infrastructure/dto/dto_scrapper"
	"LinkTracker/internal/infrastructure/httpapi"
	"LinkTracker/internal/infrastructure/httpapi/links"
	"LinkTracker/internal/infrastructure/httpapi/links/mocks"
)

func Test_GetLinksHandler_InvalidChatID(t *testing.T) {
	mockGetter := &mocks.LinkGetter{}
	handler := links.GetLinksHandler{LinkGetter: mockGetter}

	req := httptest.NewRequest(http.MethodGet, "/links", http.NoBody)
	req.Header.Set("Tg-Chat-Id", "invalid_chat_id")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	var apiError scrapperdto.ApiErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &apiError)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "INVALID_CHAT_ID", *apiError.Code)
	assert.Equal(t, "Invalid or missing chat ID", *apiError.Description)
	assert.Equal(t, "BadRequest", *apiError.ExceptionName)
}

func Test_GetLinksHandler_UserNotExist(t *testing.T) {
	mockGetter := &mocks.LinkGetter{}
	tgChatID := int64(123)
	mockGetter.On("GetUserLinks", tgChatID).Return([]domain.Link{}, domain.ErrUserNotExist{}).Once()

	handler := links.GetLinksHandler{LinkGetter: mockGetter}

	req := httptest.NewRequest(http.MethodGet, "/links", http.NoBody)
	req.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgChatID, 10))

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	var apiError scrapperdto.ApiErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &apiError)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "CHAT_NOT_EXIST", *apiError.Code)
	assert.Equal(t, "Chat not exist", *apiError.Description)
	assert.Equal(t, "BadRequest", *apiError.ExceptionName)

	mockGetter.AssertExpectations(t)
}

func Test_GetLinksHandler_LinksNotReceived(t *testing.T) {
	mockGetter := &mocks.LinkGetter{}
	tgChatID := int64(123)
	mockGetter.On("GetUserLinks", tgChatID).
		Return([]domain.Link{}, errors.New("some error")).Once()

	handler := links.GetLinksHandler{LinkGetter: mockGetter}

	req := httptest.NewRequest(http.MethodGet, "/links", http.NoBody)
	req.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgChatID, 10))

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	var apiError scrapperdto.ApiErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &apiError)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "LINKS_NOT_RECEIVED", *apiError.Code)
	assert.Equal(t, "Links not received", *apiError.Description)
	assert.Equal(t, "Server Error", *apiError.ExceptionName)

	mockGetter.AssertExpectations(t)
}

func Test_GetLinksHandler_Success(t *testing.T) {
	mockGetter := &mocks.LinkGetter{}
	tgChatID := int64(123)
	linksList := []domain.Link{
		{
			URL:     "https://example.com",
			Tags:    []string{"tag1", "tag2"},
			Filters: []string{"filter1"},
			ID:      0,
		},
	}
	mockGetter.On("GetUserLinks", tgChatID).
		Return(linksList, nil).Once()

	handler := links.GetLinksHandler{LinkGetter: mockGetter}

	req := httptest.NewRequest(http.MethodGet, "/links", http.NoBody)
	req.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgChatID, 10))

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	expectedDTO := httpapi.LinksToDTO(linksList)

	var actualResponse interface{}
	err := json.NewDecoder(rr.Body).Decode(&actualResponse)
	require.NoError(t, err)

	expectedResponse, err := json.Marshal(expectedDTO)
	require.NoError(t, err)

	var expectedResponseJSON interface{}
	err = json.Unmarshal(expectedResponse, &expectedResponseJSON)
	require.NoError(t, err)

	assert.Equal(t, expectedResponseJSON, actualResponse)

	mockGetter.AssertExpectations(t)
}
