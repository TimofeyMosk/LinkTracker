package clients_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients"
	scrapperdto "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/dto/dto_scrapper"
)

func ptrString(s string) *string {
	return &s
}

func ptrInt64(i int64) *int64 {
	return &i
}

func TestScrapperHTTPClient_RegisterUser_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/tg-chat/")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := clients.NewScrapperHTTPClient(server.URL, 2*time.Second)
	require.NoError(t, err)

	err = client.RegisterUser(12345)

	assert.NoError(t, err)
}

func TestScrapperHTTPClient_RegisterUser_BadRequest(t *testing.T) {
	errorResp := scrapperdto.ApiErrorResponse{
		Code:             ptrString("ERR_INVALID"),
		Description:      ptrString("Invalid input"),
		ExceptionMessage: ptrString("some error"),
		ExceptionName:    ptrString("BadRequest"),
		Stacktrace:       &[]string{"line1", "line2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(errorResp)
	}))
	defer server.Close()

	client, err := clients.NewScrapperHTTPClient(server.URL, 2*time.Second)
	require.NoError(t, err)

	err = client.RegisterUser(12345)

	var apiErr domain.ErrAPI

	require.Error(t, err)
	assert.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "ERR_INVALID", apiErr.Code)
	assert.Equal(t, "Invalid input", apiErr.Description)
}

func TestScrapperHTTPClient_RegisterUser_UnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := clients.NewScrapperHTTPClient(server.URL, 2*time.Second)
	require.NoError(t, err)

	err = client.RegisterUser(12345)

	var unexpected domain.ErrUnexpectedStatusCode

	require.Error(t, err)
	assert.ErrorAs(t, err, &unexpected)
	assert.Equal(t, http.StatusInternalServerError, unexpected.StatusCode)
}

func TestScrapperHTTPClient_DeleteUser_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, "/tg-chat/")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := clients.NewScrapperHTTPClient(server.URL, 2*time.Second)
	require.NoError(t, err)

	err = client.DeleteUser(12345)

	assert.NoError(t, err)
}

func TestScrapperHTTPClient_DeleteUser_NotFound(t *testing.T) {
	errorResp := scrapperdto.ApiErrorResponse{
		Code:        ptrString("ERR_NOT_FOUND"),
		Description: ptrString("Chat not exist"),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(errorResp)
	}))
	defer server.Close()

	client, err := clients.NewScrapperHTTPClient(server.URL, 2*time.Second)
	require.NoError(t, err)

	err = client.DeleteUser(12345)

	var apiErr domain.ErrAPI

	require.Error(t, err)
	assert.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "ERR_NOT_FOUND", apiErr.Code)
	assert.Equal(t, "Chat not exist", apiErr.Description)
}

func TestScrapperHTTPClient_DeleteUser_UnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := clients.NewScrapperHTTPClient(server.URL, 2*time.Second)
	require.NoError(t, err)

	err = client.DeleteUser(12345)

	var unexpected domain.ErrUnexpectedStatusCode

	require.Error(t, err)
	assert.ErrorAs(t, err, &unexpected)
	assert.Equal(t, http.StatusInternalServerError, unexpected.StatusCode)
}

func TestScrapperHTTPClient_GetLinks_Success(t *testing.T) {
	// Arrange: сервер возвращает 200 OK с валидным JSON-ответом.
	linkResponse := scrapperdto.LinkResponse{
		Url:     ptrString("https://example.com"),
		Id:      ptrInt64(1),
		Tags:    &[]string{"tag1", "tag2"},
		Filters: &[]string{"filter1"},
	}
	links := []scrapperdto.LinkResponse{linkResponse}
	size := int32(len(links)) //nolint:gosec // overflow impossible
	listResp := scrapperdto.ListLinksResponse{
		Links: &links,
		Size:  &size,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "12345", r.Header.Get("Tg-Chat-Id"))
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(listResp)
	}))
	defer server.Close()

	client, err := clients.NewScrapperHTTPClient(server.URL, 2*time.Second)
	require.NoError(t, err)

	// Act
	result, err := client.GetLinks(12345)

	// Assert
	require.NoError(t, err)

	expected := []domain.Link{{
		URL:     "https://example.com",
		Tags:    []string{"tag1", "tag2"},
		Filters: []string{"filter1"},
		ID:      1,
	}}
	assert.Equal(t, expected, result)
}

func TestScrapperHTTPClient_GetLinks_BadRequest(t *testing.T) {
	// Arrange: сервер возвращает 400 с ошибочным JSON-ответом.
	errorResp := scrapperdto.ApiErrorResponse{
		Code:        ptrString("ERR_BAD_REQUEST"),
		Description: ptrString("Invalid input"),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(errorResp)
	}))
	defer server.Close()

	client, err := clients.NewScrapperHTTPClient(server.URL, 2*time.Second)
	require.NoError(t, err)

	// Act
	result, err := client.GetLinks(12345)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)

	var apiErr domain.ErrAPI

	require.Error(t, err)
	assert.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "ERR_BAD_REQUEST", apiErr.Code)
	assert.Equal(t, "Invalid input", apiErr.Description)
}

func TestScrapperHTTPClient_GetLinks_UnexpectedStatus(t *testing.T) {
	// Arrange: сервер возвращает неожиданный статус (500)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := clients.NewScrapperHTTPClient(server.URL, 2*time.Second)
	require.NoError(t, err)

	result, err := client.GetLinks(12345)

	require.Error(t, err)
	assert.Nil(t, result)

	var unexpected domain.ErrUnexpectedStatusCode

	require.Error(t, err)
	assert.ErrorAs(t, err, &unexpected)
	assert.Equal(t, http.StatusInternalServerError, unexpected.StatusCode)
}

func TestScrapperHTTPClient_AddLink_Success(t *testing.T) {
	linkResponse := scrapperdto.LinkResponse{
		Url:     ptrString("https://example.com"),
		Id:      ptrInt64(1),
		Tags:    &[]string{"tag1", "tag2"},
		Filters: &[]string{"filter1"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "12345", r.Header.Get("Tg-Chat-Id"))
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(linkResponse)
	}))
	defer server.Close()

	client, err := clients.NewScrapperHTTPClient(server.URL, 2*time.Second)
	require.NoError(t, err)

	testLink := domain.Link{
		URL:     "https://example.com",
		Tags:    []string{"tag1", "tag2"},
		Filters: []string{"filter1"},
		ID:      1,
	}

	err = client.AddLink(12345, testLink)

	assert.NoError(t, err)
}

func TestScrapperHTTPClient_AddLink_BadRequest(t *testing.T) {
	errorResp := scrapperdto.ApiErrorResponse{
		Code:        ptrString("ERR_ADD_LINK"),
		Description: ptrString("Failed to add link"),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(errorResp)
	}))
	defer server.Close()

	client, err := clients.NewScrapperHTTPClient(server.URL, 2*time.Second)
	require.NoError(t, err)

	testLink := domain.Link{
		URL:     "https://example.com",
		Tags:    []string{"tag1", "tag2"},
		Filters: []string{"filter1"},
		ID:      1,
	}

	err = client.AddLink(12345, testLink)

	var apiErr domain.ErrAPI

	require.Error(t, err)
	assert.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "ERR_ADD_LINK", apiErr.Code)
	assert.Equal(t, "Failed to add link", apiErr.Description)
}

func TestScrapperHTTPClient_AddLink_UnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := clients.NewScrapperHTTPClient(server.URL, 2*time.Second)
	require.NoError(t, err)

	testLink := domain.Link{
		URL:     "https://example.com",
		Tags:    []string{"tag1", "tag2"},
		Filters: []string{"filter1"},
		ID:      1,
	}

	err = client.AddLink(12345, testLink)

	var unexpected domain.ErrUnexpectedStatusCode

	require.Error(t, err)
	assert.ErrorAs(t, err, &unexpected)
	assert.Equal(t, http.StatusInternalServerError, unexpected.StatusCode)
}

func TestScrapperHTTPClient_RemoveLink_Success(t *testing.T) {
	linkResponse := scrapperdto.LinkResponse{
		Url:     ptrString("https://example.com"),
		Id:      ptrInt64(1),
		Tags:    &[]string{"tag1", "tag2"},
		Filters: &[]string{"filter1"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "12345", r.Header.Get("Tg-Chat-Id"))
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(linkResponse)
	}))
	defer server.Close()

	client, err := clients.NewScrapperHTTPClient(server.URL, 2*time.Second)
	require.NoError(t, err)

	testLink := domain.Link{
		URL:     "https://example.com",
		Tags:    []string{"tag1", "tag2"},
		Filters: []string{"filter1"},
		ID:      1,
	}

	err = client.RemoveLink(12345, testLink)

	assert.NoError(t, err)
}

func TestScrapperHTTPClient_RemoveLink_BadRequest(t *testing.T) {
	errorResp := scrapperdto.ApiErrorResponse{
		Code:        ptrString("ERR_REMOVE_LINK"),
		Description: ptrString("Failed to remove link"),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(errorResp)
	}))
	defer server.Close()

	client, err := clients.NewScrapperHTTPClient(server.URL, 2*time.Second)
	require.NoError(t, err)

	testLink := domain.Link{
		URL:     "https://example.com",
		Tags:    []string{"tag1", "tag2"},
		Filters: []string{"filter1"},
		ID:      1,
	}

	err = client.RemoveLink(12345, testLink)

	var apiErr domain.ErrAPI

	require.Error(t, err)
	assert.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "ERR_REMOVE_LINK", apiErr.Code)
	assert.Equal(t, "Failed to remove link", apiErr.Description)
}

func TestScrapperHTTPClient_RemoveLink_UnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := clients.NewScrapperHTTPClient(server.URL, 2*time.Second)
	require.NoError(t, err)

	testLink := domain.Link{
		URL:     "https://example.com",
		Tags:    []string{"tag1", "tag2"},
		Filters: []string{"filter1"},
		ID:      1,
	}

	err = client.RemoveLink(12345, testLink)

	var unexpected domain.ErrUnexpectedStatusCode

	require.Error(t, err)
	assert.ErrorAs(t, err, &unexpected)
	assert.Equal(t, http.StatusInternalServerError, unexpected.StatusCode)
}
