package clients_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"LinkTracker/internal/infrastructure/clients"
)

// roundTripGitFunc позволяет использовать функцию как http.RoundTripper.
type roundTripGitFunc func(req *http.Request) (*http.Response, error)

func (f roundTripGitFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// newTestGitHubHTTPClient подменяет URL запросов на тестовый сервер.
func newTestGitHubHTTPClient(testServerURL string, _ time.Duration, rt roundTripGitFunc) *clients.GitHubHTTPClient {
	client := clients.NewGitHubHTTPClient()
	client.Client.Transport = roundTripGitFunc(func(req *http.Request) (*http.Response, error) {
		// Если URL начинается с реального API, заменяем его на testServerURL.
		if strings.HasPrefix(req.URL.String(), "https://api.github.com") {
			parsed, err := url.Parse(testServerURL)
			if err != nil {
				return nil, err
			}

			req.URL.Scheme = parsed.Scheme
			req.URL.Host = parsed.Host
		}

		return rt(req)
	})

	return client
}

func TestGitHubHTTPClient_GetLatestPROrIssue(t *testing.T) {
	tests := []struct {
		name         string
		responseBody string
		expectError  bool
		expectedDesc string
		expectedTime time.Time
	}{
		{
			name: "Valid issue",
			responseBody: `[
				{
					"title": "Test Issue",
					"user": {"login": "testuser"},
					"body": "This is a test issue body",
					"created_at": "2020-01-01T12:00:00Z"
				}
			]`,
			expectError:  false,
			expectedDesc: "Title: Test Issue\nUser: testuser\nCreated At: 2020-01-01T12:00:00Z\nPreview: This is a test issue body",
			expectedTime: time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name:         "Empty issues array",
			responseBody: `[]`,
			expectError:  true,
		},
		{
			name:         "Invalid JSON",
			responseBody: `invalid json`,
			expectError:  true,
		},
		{
			name: "Invalid time format",
			responseBody: `[
				{
					"title": "Issue with bad time",
					"user": {"login": "user1"},
					"body": "Body text",
					"created_at": "not-a-time"
				}
			]`,
			expectError: true,
		},
	}

	testServerURL := "http://example.com"

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rt := roundTripGitFunc(func(_ *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(tc.responseBody)),
					Header:     make(http.Header),
				}, nil
			})

			client := newTestGitHubHTTPClient(testServerURL, 5*time.Second, rt)
			lastUpdate, description, err := client.GetLatestPROrIssue(context.Background(), "https://github.com/owner/repo")

			if tc.expectError {
				assert.Error(t, err, "expected an error but got none")
				return
			}

			assert.NoError(t, err, "unexpected error occurred")
			assert.Equal(t, tc.expectedTime, lastUpdate, "unexpected timestamp")
			assert.Equal(t, tc.expectedDesc, description, "unexpected description")
		})
	}
}
