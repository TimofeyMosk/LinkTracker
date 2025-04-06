package clients_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"LinkTracker/internal/infrastructure/clients"
)

// roundTripSOFunc используется для подмены транспорта HTTP-клиента.
type roundTripSOFunc func(req *http.Request) (*http.Response, error)

func (f roundTripSOFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// testQuestionLink — тестовая ссылка на вопрос.
const testQuestionLink = "https://stackoverflow.com/questions/12345/some-question-title"

// newTestClient создаёт поддельный StackOverflowHTTPClient с кастомным RoundTripper.
func newTestClient(rt roundTripSOFunc) *clients.StackOverflowHTTPClient {
	client := clients.NewStackOverflowHTTPClient()
	client.Client.Transport = rt

	return client
}

func TestStackOverflowHTTPClient_GetLatestAnswerOrComment(t *testing.T) {
	tests := []struct {
		name               string
		questionResponse   string
		answerResponse     string
		commentResponse    string
		expectError        bool
		expectedDescSubstr string
		expectedTime       time.Time
	}{
		{
			name: "Valid answer exists",
			questionResponse: `{
				"items": [{"title": "Test Question"}]
			}`,
			answerResponse: `{
				"items": [{
					"owner": {"display_name": "AnswerUser"},
					"creation_date": 1580505600,
					"body": "This is the answer body"
				}]
			}`,
			commentResponse:    `{"items": []}`,
			expectError:        false,
			expectedDescSubstr: "Question: Test Question",
			expectedTime:       time.Unix(1580505600, 0),
		},
		{
			name: "No answer, valid comment exists",
			questionResponse: `{
				"items": [{"title": "Test Question 2"}]
			}`,
			answerResponse: `{"items": []}`,
			commentResponse: `{
				"items": [{
					"owner": {"display_name": "CommentUser"},
					"creation_date": 1609459200,
					"body": "This is a comment body"
				}]
			}`,
			expectError:        false,
			expectedDescSubstr: "Question: Test Question 2",
			expectedTime:       time.Unix(1609459200, 0),
		},
		{
			name: "No answer and no comment",
			questionResponse: `{
				"items": [{"title": "Test Question 3"}]
			}`,
			answerResponse:  `{"items": []}`,
			commentResponse: `{"items": []}`,
			expectError:     true,
		},
		{
			name:             "Invalid question details",
			questionResponse: `invalid json`,
			answerResponse:   `{"items": []}`,
			commentResponse:  `{"items": []}`,
			expectError:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rt := roundTripSOFunc(func(req *http.Request) (*http.Response, error) {
				var bodyStr string

				switch {
				case strings.Contains(req.URL.Path, "/questions/12345") && !strings.Contains(req.URL.Path, "/answers") &&
					!strings.Contains(req.URL.Path, "/comments"):
					bodyStr = tc.questionResponse
				case strings.Contains(req.URL.Path, "/questions/12345/answers"):
					bodyStr = tc.answerResponse
				case strings.Contains(req.URL.Path, "/questions/12345/comments"):
					bodyStr = tc.commentResponse
				default:
					return nil, fmt.Errorf("unexpected request: %s", req.URL.Path)
				}

				resp := &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(bodyStr)),
					Header:     make(http.Header),
				}

				return resp, nil
			})

			client := newTestClient(rt)
			lastUpdate, description, err := client.GetLatestAnswerOrComment(context.Background(), testQuestionLink)

			if tc.expectError {
				assert.Error(t, err, "expected an error but got none")
				return
			}

			assert.NoError(t, err, "unexpected error occurred")
			assert.Equal(t, tc.expectedTime, lastUpdate, "unexpected timestamp")
			assert.Contains(t, description, tc.expectedDescSubstr, "description does not contain expected substring")
		})
	}
}
