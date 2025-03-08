package clients_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients"
)

func TestBotHTTPClient_PostUpdates_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/updates", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := clients.NewBotHTTPClient(server.URL, 2*time.Second)
	assert.NoError(t, err)

	link := domain.Link{ID: 1, URL: "https://example.com"}
	err = client.PostUpdates(link, 123456)
	assert.NoError(t, err)
}

func TestBotHTTPClient_PostUpdates_BadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code": "INVALID_INPUT", "description": "Invalid data"}`))
	}))
	defer server.Close()

	client, err := clients.NewBotHTTPClient(server.URL, 2*time.Second)
	assert.NoError(t, err)

	link := domain.Link{ID: 1, URL: "https://example.com"}
	err = client.PostUpdates(link, 123456)
	assert.Error(t, err)

	var apiErr domain.ErrAPI
	ok := errors.As(err, &apiErr)
	assert.True(t, ok)

	assert.Equal(t, "INVALID_INPUT", apiErr.Code)
	assert.Equal(t, "Invalid data", apiErr.Description)
}

func TestBotHTTPClient_PostUpdates_UnexpectedStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := clients.NewBotHTTPClient(server.URL, 2*time.Second)
	assert.NoError(t, err)

	link := domain.Link{ID: 1, URL: "https://example.com"}
	err = client.PostUpdates(link, 123456)
	assert.Error(t, err)

	var errUnexpectedStatusCode domain.ErrUnexpectedStatusCode
	ok := errors.As(err, &errUnexpectedStatusCode)
	assert.True(t, ok)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, errUnexpectedStatusCode.StatusCode)
}
