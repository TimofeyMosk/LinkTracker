package clients_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"LinkTracker/internal/infrastructure/clients"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const gitOwnerRepo = "https://github.com/owner/repo"

func newTestGitHubHTTPClient(testServerURL string, _ time.Duration) *clients.GitHubHTTPClient {
	client := clients.NewGitHubHTTPClient()
	client.Client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if strings.HasPrefix(req.URL.String(), "https://api.github.com") {
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

func Test_GitHubHTTPClient_GetLastUpdateTimeRepo_Success(t *testing.T) {
	expectedTimeStr := "2022-01-01T12:00:00Z"
	expectedTime, err := time.Parse(time.RFC3339, expectedTimeStr)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/owner/repo", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"updated_at": expectedTimeStr,
		})
	}))
	defer ts.Close()

	client := newTestGitHubHTTPClient(ts.URL, 5*time.Second)
	link := gitOwnerRepo

	result, err := client.GetLastUpdateTimeRepo(link)

	require.NoError(t, err)
	assert.Equal(t, expectedTime, result)
}

func Test_GitHubHTTPClient_GetLastUpdateTimeRepo_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer ts.Close()

	client := newTestGitHubHTTPClient(ts.URL, 5*time.Second)
	link := gitOwnerRepo

	_, err := client.GetLastUpdateTimeRepo(link)

	require.Error(t, err)
}

func Test_GitHubHTTPClient_GetLastUpdateTimeRepo_InvalidTimeFormat(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"updated_at": "not-a-time",
		})
	}))
	defer ts.Close()

	client := newTestGitHubHTTPClient(ts.URL, 5*time.Second)
	link := gitOwnerRepo

	_, err := client.GetLastUpdateTimeRepo(link)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing time")
}

func Test_GitHubHTTPClient_GetLastUpdateTimeRepo_InvalidLink(t *testing.T) {
	client := clients.NewGitHubHTTPClient()
	invalidLink := "https://github.com/owner" // отсутствует имя репозитория

	_, err := client.GetLastUpdateTimeRepo(invalidLink)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "wrong url format")
}
