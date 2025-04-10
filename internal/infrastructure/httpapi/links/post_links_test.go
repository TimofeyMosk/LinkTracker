package links_test

//
//import (
//	"bytes"
//	"encoding/json"
//	"errors"
//	"net/http"
//	"net/http/httptest"
//	"strconv"
//	"strings"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//
//	"LinkTracker/internal/domain"
//	scrapperdto "LinkTracker/internal/infrastructure/dto/dto_scrapper"
//	"LinkTracker/internal/infrastructure/httpapi/links"
//	"LinkTracker/internal/infrastructure/httpapi/links/mocks"
//)
//
//func ptr(s string) *string {
//	return &s
//}
//
//func Test_PostLinkHandler_InvalidChatID(t *testing.T) {
//	mockAdder := &mocks.LinkAdder{}
//	handler := links.PostLinksHandler{LinkAdder: mockAdder}
//
//	req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewBufferString(`{}`))
//	req.Header.Set("Tg-Chat-Id", "invalid_chat_id")
//
//	rr := httptest.NewRecorder()
//
//	handler.ServeHTTP(rr, req)
//
//	var apiError scrapperdto.ApiErrorResponse
//	err := json.Unmarshal(rr.Body.Bytes(), &apiError)
//	require.NoError(t, err)
//
//	assert.Equal(t, http.StatusBadRequest, rr.Code)
//	assert.Equal(t, "INVALID_CHAT_ID", *apiError.Code)
//	assert.Equal(t, "Invalid or missing chat ID", *apiError.Description)
//	assert.Equal(t, "BadRequest", *apiError.ExceptionName)
//}
//
//func Test_PostLinkHandler_InvalidRequestBody_JSON(t *testing.T) {
//	mockAdder := &mocks.LinkAdder{}
//	handler := links.PostLinksHandler{LinkAdder: mockAdder}
//
//	req := httptest.NewRequest(http.MethodPost, "/links", strings.NewReader("invalid json"))
//	req.Header.Set("Tg-Chat-Id", "123")
//
//	rr := httptest.NewRecorder()
//
//	handler.ServeHTTP(rr, req)
//
//	var apiError scrapperdto.ApiErrorResponse
//	err := json.Unmarshal(rr.Body.Bytes(), &apiError)
//	require.NoError(t, err)
//
//	assert.Equal(t, http.StatusBadRequest, rr.Code)
//	assert.Equal(t, "INVALID_REQUEST_BODY", *apiError.Code)
//	assert.Equal(t, "Invalid or missing request body", *apiError.Description)
//	assert.Equal(t, "BadRequest", *apiError.ExceptionName)
//}
//
//func Test_PostLinkHandler_InvalidRequestBody_Dto(t *testing.T) {
//	mockAdder := &mocks.LinkAdder{}
//	handler := links.PostLinksHandler{LinkAdder: mockAdder}
//
//	reqBodyStruct := scrapperdto.LinkRequest{
//		Link:    nil,
//		Tags:    &[]string{"tag1"},
//		Filters: &[]string{"filter1"},
//	}
//	reqBody, err := json.Marshal(reqBodyStruct)
//	require.NoError(t, err)
//
//	req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewReader(reqBody))
//	req.Header.Set("Tg-Chat-Id", "123")
//
//	rr := httptest.NewRecorder()
//
//	handler.ServeHTTP(rr, req)
//
//	var apiError scrapperdto.ApiErrorResponse
//	err = json.Unmarshal(rr.Body.Bytes(), &apiError)
//	require.NoError(t, err)
//
//	assert.Equal(t, http.StatusBadRequest, rr.Code)
//	assert.Equal(t, "INVALID_REQUEST_BODY", *apiError.Code)
//	assert.Equal(t, "Invalid or missing request body", *apiError.Description)
//	assert.Equal(t, "BadRequest", *apiError.ExceptionName)
//}
//
//func Test_PostLinkHandler_ChatNotExist(t *testing.T) {
//	mockAdder := &mocks.LinkAdder{}
//	tgChatID := int64(123)
//	linkInput := domain.Link{
//		URL:     "https://example.com",
//		Tags:    []string{"tag1"},
//		Filters: []string{"filter1"},
//	}
//	mockAdder.On("AddLink", tgChatID, linkInput).
//		Return(domain.Link{}, domain.ErrUserNotExist{}).Once()
//
//	handler := links.PostLinksHandler{LinkAdder: mockAdder}
//
//	reqBodyStruct := scrapperdto.LinkRequest{
//		Link:    ptr("https://example.com"),
//		Tags:    &[]string{"tag1"},
//		Filters: &[]string{"filter1"},
//	}
//	reqBody, err := json.Marshal(reqBodyStruct)
//	require.NoError(t, err)
//
//	req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewReader(reqBody))
//	req.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgChatID, 10))
//
//	rr := httptest.NewRecorder()
//
//	handler.ServeHTTP(rr, req)
//
//	var apiError scrapperdto.ApiErrorResponse
//	err = json.Unmarshal(rr.Body.Bytes(), &apiError)
//	require.NoError(t, err)
//
//	assert.Equal(t, http.StatusBadRequest, rr.Code)
//	assert.Equal(t, "CHAT_NOT_EXIST", *apiError.Code)
//	assert.Equal(t, "Chat not exist", *apiError.Description)
//	assert.Equal(t, "BadRequest", *apiError.ExceptionName)
//
//	mockAdder.AssertExpectations(t)
//}
//
//func Test_PostLinkHandler_LinkAlreadyTracking(t *testing.T) {
//	mockAdder := &mocks.LinkAdder{}
//	tgChatID := int64(123)
//	linkInput := domain.Link{
//		URL:     "https://example.com",
//		Tags:    []string{"tag1"},
//		Filters: []string{"filter1"},
//	}
//	// Симулируем ошибку "ссылка уже отслеживается".
//	mockAdder.On("AddLink", tgChatID, linkInput).
//		Return(domain.Link{}, domain.ErrLinkAlreadyTracking{}).Once()
//
//	handler := links.PostLinksHandler{LinkAdder: mockAdder}
//
//	reqBodyStruct := scrapperdto.LinkRequest{
//		Link:    ptr("https://example.com"),
//		Tags:    &[]string{"tag1"},
//		Filters: &[]string{"filter1"},
//	}
//	reqBody, err := json.Marshal(reqBodyStruct)
//	require.NoError(t, err)
//
//	req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewReader(reqBody))
//	req.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgChatID, 10))
//
//	rr := httptest.NewRecorder()
//
//	handler.ServeHTTP(rr, req)
//
//	var apiError scrapperdto.ApiErrorResponse
//	err = json.Unmarshal(rr.Body.Bytes(), &apiError)
//	require.NoError(t, err)
//
//	assert.Equal(t, http.StatusBadRequest, rr.Code)
//	assert.Equal(t, "LINKS_ALREADY_EXIST", *apiError.Code)
//	assert.Equal(t, "Links already tracking", *apiError.Description)
//	assert.Equal(t, "BadRequest", *apiError.ExceptionName)
//
//	mockAdder.AssertExpectations(t)
//}
//
//func Test_PostLinkHandler_AddLinkFailed(t *testing.T) {
//	mockAdder := &mocks.LinkAdder{}
//	tgChatID := int64(123)
//	linkInput := domain.Link{
//		URL:     "https://example.com",
//		Tags:    []string{"tag1"},
//		Filters: []string{"filter1"},
//	}
//	errMsg := "some error"
//	mockAdder.On("AddLink", tgChatID, linkInput).
//		Return(domain.Link{}, errors.New(errMsg)).Once()
//
//	handler := links.PostLinksHandler{LinkAdder: mockAdder}
//
//	reqBodyStruct := scrapperdto.LinkRequest{
//		Link:    ptr("https://example.com"),
//		Tags:    &[]string{"tag1"},
//		Filters: &[]string{"filter1"},
//	}
//	reqBody, err := json.Marshal(reqBodyStruct)
//	require.NoError(t, err)
//
//	req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewReader(reqBody))
//	req.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgChatID, 10))
//
//	rr := httptest.NewRecorder()
//
//	handler.ServeHTTP(rr, req)
//
//	var apiError scrapperdto.ApiErrorResponse
//	err = json.Unmarshal(rr.Body.Bytes(), &apiError)
//	require.NoError(t, err)
//
//	assert.Equal(t, http.StatusBadRequest, rr.Code)
//	assert.Equal(t, "ADD_LINK_FAILED", *apiError.Code)
//	assert.Equal(t, "Failed to added link", *apiError.Description)
//	assert.Equal(t, "Server Error", *apiError.ExceptionName)
//
//	mockAdder.AssertExpectations(t)
//}
//
//func Test_PostLinkHandler_Success(t *testing.T) {
//	mockAdder := &mocks.LinkAdder{}
//	tgChatID := int64(123)
//	inputLink := domain.Link{
//		URL:     "https://example.com",
//		Tags:    []string{"tag1"},
//		Filters: []string{"filter1"},
//	}
//	returnedLink := domain.Link{
//		ID:      456,
//		URL:     "https://example.com",
//		Tags:    []string{"tag1"},
//		Filters: []string{"filter1"},
//	}
//	mockAdder.On("AddLink", tgChatID, inputLink).
//		Return(returnedLink, nil).Once()
//
//	handler := links.PostLinksHandler{LinkAdder: mockAdder}
//
//	reqBodyStruct := scrapperdto.LinkRequest{
//		Link:    ptr("https://example.com"),
//		Tags:    &[]string{"tag1"},
//		Filters: &[]string{"filter1"},
//	}
//	reqBody, err := json.Marshal(reqBodyStruct)
//	require.NoError(t, err)
//
//	req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewReader(reqBody))
//	req.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgChatID, 10))
//
//	rr := httptest.NewRecorder()
//
//	handler.ServeHTTP(rr, req)
//
//	assert.Equal(t, http.StatusOK, rr.Code)
//	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
//
//	var linkResp scrapperdto.LinkResponse
//	err = json.NewDecoder(rr.Body).Decode(&linkResp)
//	require.NoError(t, err)
//
//	assert.NotNil(t, linkResp.Id)
//	assert.Equal(t, returnedLink.ID, *linkResp.Id)
//	assert.NotNil(t, linkResp.Url)
//	assert.Equal(t, returnedLink.URL, *linkResp.Url)
//	assert.NotNil(t, linkResp.Tags)
//	assert.Equal(t, returnedLink.Tags, *linkResp.Tags)
//	assert.NotNil(t, linkResp.Filters)
//	assert.Equal(t, returnedLink.Filters, *linkResp.Filters)
//
//	mockAdder.AssertExpectations(t)
//}
