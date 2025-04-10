package links_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	scrapperdto "LinkTracker/internal/infrastructure/dto/dto_scrapper"
	"LinkTracker/internal/infrastructure/httpapi/links"
	"LinkTracker/internal/infrastructure/httpapi/links/mocks"
)

// const GitUserRepo = "https://github.com/user/repo"
func Test_DeleteLinksHandler_InvalidChatID(t *testing.T) {
	mockScrapper := &mocks.LinkDeleter{}
	handler := links.DeleteLinksHandler{LinkDeleter: mockScrapper}
	request := httptest.NewRequest(http.MethodPost, "/links", http.NoBody)
	request.Header.Set("Tg-Chat-Id", "invalid_chat_id")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	var apiErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &apiErrorBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	assert.Equal(t, "INVALID_CHAT_ID", *apiErrorBody.Code)
	assert.Equal(t, "Invalid or missing chat ID", *apiErrorBody.Description)
	assert.Equal(t, "BadRequest", *apiErrorBody.ExceptionName)
}

func Test_DeleteLinksHandler_InvalidRequestBody(t *testing.T) {
	mockScrapper := &mocks.LinkDeleter{}
	handler := links.DeleteLinksHandler{LinkDeleter: mockScrapper}

	request := httptest.NewRequest(http.MethodPost, "/links", strings.NewReader("invalid json"))
	request.Header.Set("Tg-Chat-Id", "123")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	var apiErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &apiErrorBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	assert.Equal(t, "INVALID_REQUEST_BODY", *apiErrorBody.Code)
	assert.Equal(t, "Invalid or missing request body", *apiErrorBody.Description)
	assert.Equal(t, "BadRequest", *apiErrorBody.ExceptionName)
}

//func Test_DeleteLinksHandler_UserNotExist(t *testing.T) {
//	mockScrapper := &mocks.LinkDeleter{}
//	tgChatID := int64(123)
//	linkURL := "https://example.com"
//	link := domain.Link{URL: linkURL, Tags: nil, Filters: nil, ID: 0}
//	mockScrapper.On("DeleteLink", tgChatID, link).
//		Return(domain.Link{}, domain.ErrUserNotExist{}).Once()
//
//	handler := links.DeleteLinksHandler{LinkDeleter: mockScrapper}
//
//	removeLinkRequestBody := scrapperdto.RemoveLinkRequest{Link: &linkURL}
//	reqBody, err := json.Marshal(removeLinkRequestBody)
//	require.NoError(t, err)
//
//	request := httptest.NewRequest(http.MethodDelete, "/links", bytes.NewReader(reqBody))
//	request.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgChatID, 10))
//
//	responseRecorder := httptest.NewRecorder()
//
//	handler.ServeHTTP(responseRecorder, request)
//
//	var apiErrorBody scrapperdto.ApiErrorResponse
//	err = json.Unmarshal(responseRecorder.Body.Bytes(), &apiErrorBody)
//	require.NoError(t, err)
//
//	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)
//	assert.Equal(t, "CHAT_NOT_EXIST", *apiErrorBody.Code)
//	assert.Equal(t, "Chat not exist", *apiErrorBody.Description)
//	assert.Equal(t, "Not Found", *apiErrorBody.ExceptionName)
//	mockScrapper.AssertExpectations(t)
//}

//func Test_DeleteLinksHandler_DeleteLinkFailed(t *testing.T) {
//	mockScrapper := &mocks.LinkDeleter{}
//	tgChatID := int64(123)
//	linkURL := "https://example.com"
//	link := domain.Link{URL: linkURL, Tags: nil, Filters: nil, ID: 0}
//	mockScrapper.On("DeleteLink", tgChatID, link).
//		Return(domain.Link{}, errors.New("some error")).Once()
//
//	handler := links.DeleteLinksHandler{LinkDeleter: mockScrapper}
//
//	removeLinkRequestBody := scrapperdto.RemoveLinkRequest{Link: &linkURL}
//	reqBody, err := json.Marshal(removeLinkRequestBody)
//	require.NoError(t, err)
//
//	request := httptest.NewRequest(http.MethodDelete, "/links", bytes.NewReader(reqBody))
//	request.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgChatID, 10))
//
//	responseRecorder := httptest.NewRecorder()
//
//	handler.ServeHTTP(responseRecorder, request)
//
//	var apiErrorBody scrapperdto.ApiErrorResponse
//	err = json.Unmarshal(responseRecorder.Body.Bytes(), &apiErrorBody)
//	require.NoError(t, err)
//
//	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
//	assert.Equal(t, "DELETE_LINK_FAILED", *apiErrorBody.Code)
//	assert.Equal(t, "Failed to delete link", *apiErrorBody.Description)
//	assert.Equal(t, "BadRequest", *apiErrorBody.ExceptionName)
//	mockScrapper.AssertExpectations(t)
//}
//
//func Test_DeleteLinksHandler_Success(t *testing.T) {
//	mockScrapper := &mocks.LinkDeleter{}
//	tgChatID := int64(123)
//	expectedLink := domain.Link{
//		URL:     "https://github.com/example/example",
//		Tags:    []string{"tag1", "tag2"},
//		Filters: []string{"filter1"},
//		ID:      0,
//	}
//
//	mockScrapper.On("DeleteLink", tgChatID, domain.Link{
//		URL: "https://github.com/example/example", Tags: nil, Filters: nil, ID: 0}).
//		Return(expectedLink, nil).Once()
//
//	handler := links.DeleteLinksHandler{LinkDeleter: mockScrapper}
//
//	linkURL := expectedLink.URL
//	removeReq := scrapperdto.RemoveLinkRequest{Link: &linkURL}
//	reqBody, err := json.Marshal(removeReq)
//	require.NoError(t, err)
//
//	request := httptest.NewRequest(http.MethodPost, "/tg-chat", bytes.NewReader(reqBody))
//	request.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgChatID, 10))
//
//	responseRecorder := httptest.NewRecorder()
//
//	handler.ServeHTTP(responseRecorder, request)
//
//	assert.Equal(t, http.StatusOK, responseRecorder.Code)
//	assert.Equal(t, "application/json", responseRecorder.Header().Get("Content-Type"))
//
//	var response scrapperdto.LinkResponse
//	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
//	require.NoError(t, err)
//
//	assert.NotNil(t, response.Url)
//	assert.Equal(t, expectedLink.URL, *response.Url)
//	assert.NotNil(t, response.Id)
//	assert.Equal(t, tgChatID, *response.Id)
//	assert.NotNil(t, response.Tags)
//	assert.Equal(t, expectedLink.Tags, *response.Tags)
//	assert.NotNil(t, response.Filters)
//	assert.Equal(t, expectedLink.Filters, *response.Filters)
//
//	mockScrapper.AssertExpectations(t)
//}
