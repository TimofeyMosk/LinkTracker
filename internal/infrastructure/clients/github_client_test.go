package clients_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients"
)

const gitOwnerRepo = "https://github.com/owner/repo"

// newTestGitHubHTTPClient создаёт клиент, который перенаправляет запросы, начинающиеся с githubAPIBaseURL, на testServer.
func newTestGitHubHTTPClient(testServerURL string, _ time.Duration) *clients.GitHubHTTPClient {
	client := clients.NewGitHubHTTPClient()
	client.Client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if strings.HasPrefix(req.URL.String(), "https://api.github.com") {
			// Заменяем схему и host на значения из testServerURL.
			parsed, err := url.Parse(testServerURL)
			if err != nil {
				return nil, err
			}

			req.URL.Scheme = parsed.Scheme
			req.URL.Host = parsed.Host
		}

		return http.DefaultTransport.RoundTrip(req)
	})

	return client
}

func TestGitHubHTTPClient_GetLastUpdateTimeRepo_Success(t *testing.T) {
	// Arrange: сервер возвращает корректный JSON с полем updated_at.
	expectedTimeStr := "2022-01-01T12:00:00Z"
	expectedTime, err := time.Parse(time.RFC3339, expectedTimeStr)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что путь соответствует формату /repos/{owner}/{repo}
		assert.Equal(t, "/repos/owner/repo", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"updated_at": expectedTimeStr,
		})
	}))
	defer ts.Close()

	client := newTestGitHubHTTPClient(ts.URL, 5*time.Second)
	link := gitOwnerRepo

	// Act
	result, err := client.GetLastUpdateTimeRepo(link)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedTime, result)
}

func TestGitHubHTTPClient_GetLastUpdateTimeRepo_InvalidJSON(t *testing.T) {
	// Arrange: сервер возвращает некорректный JSON.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer ts.Close()

	client := newTestGitHubHTTPClient(ts.URL, 5*time.Second)
	link := gitOwnerRepo

	// Act
	_, err := client.GetLastUpdateTimeRepo(link)

	// Assert
	require.Error(t, err)
}

func TestGitHubHTTPClient_GetLastUpdateTimeRepo_InvalidTimeFormat(t *testing.T) {
	// Arrange: сервер возвращает JSON с неверным форматом времени.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"updated_at": "not-a-time",
		})
	}))
	defer ts.Close()

	client := newTestGitHubHTTPClient(ts.URL, 5*time.Second)
	link := gitOwnerRepo

	// Act
	_, err := client.GetLastUpdateTimeRepo(link)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing time")
}

func TestGitHubHTTPClient_GetLastUpdateTimeRepo_InvalidLink(t *testing.T) {
	// Arrange: передаём ссылку, которая не соответствует формату (ожидается github.com/{owner}/{repo}).
	client := clients.NewGitHubHTTPClient()   // ошибка возникает до HTTP-запроса
	invalidLink := "https://github.com/owner" // отсутствует имя репозитория

	// Act
	_, err := client.GetLastUpdateTimeRepo(invalidLink)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "wrong url format")
}
