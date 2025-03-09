package clients_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// newTestClient создаёт экземпляр StackOverflowHTTPClient, чей транспорт перенаправляет запросы на testServer.
func newTestClient(testServerURL string, timeout time.Duration) *clients.StackOverflowHTTPClient {
	client := clients.NewStackOverflowHTTPClient()
	// Переопределяем Transport, чтобы перенаправлять запросы, начинающиеся с реального API-адреса,
	// на адрес нашего тестового сервера.
	client.Client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		// если URL начинается с "https://api.stackexchange.com/2.2", заменяем на testServerURL
		if strings.HasPrefix(req.URL.String(), "https://api.stackexchange.com/2.2") {
			// Изменяем схему и host запроса на значения из testServerURL.
			parsed, err := http.NewRequest(req.Method, testServerURL, http.NoBody)
			if err != nil {
				return nil, err
			}

			req.URL.Scheme = parsed.URL.Scheme
			req.URL.Host = parsed.URL.Host
		}

		return http.DefaultTransport.RoundTrip(req)
	})

	return client
}

func TestStackOverflowHTTPClient_GetLastActivityQuestion_Success(t *testing.T) {
	// Arrange: тестовый сервер возвращает корректный JSON с одним элементом.
	expectedTimestamp := time.Now().Unix()
	soResponse := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"last_activity_date": expectedTimestamp,
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что путь запроса сформирован корректно: должен содержать /questions/{id}
		assert.Contains(t, r.URL.Path, "/questions/12345")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(soResponse)
	}))
	defer ts.Close()

	client := newTestClient(ts.URL, 5*time.Second)
	// Используем корректную ссылку StackOverflow.
	link := "https://stackoverflow.com/questions/12345/some-question-title"

	// Act
	lastActivity, err := client.GetLastActivityQuestion(link)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, time.Unix(expectedTimestamp, 0), lastActivity)
}

func TestStackOverflowHTTPClient_GetLastActivityQuestion_NoItems(t *testing.T) {
	// Arrange: сервер возвращает JSON с пустым items.
	soResponse := map[string]interface{}{
		"items": []interface{}{},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(soResponse)
	}))
	defer ts.Close()

	client := newTestClient(ts.URL, 5*time.Second)
	link := "https://stackoverflow.com/questions/12345/some-question-title"

	// Act
	_, err := client.GetLastActivityQuestion(link)

	// Assert: ожидается ошибка ErrQuestionNotFound.
	require.Error(t, err)
	_, ok := err.(domain.ErrQuestionNotFound)
	assert.True(t, ok)
}

func TestStackOverflowHTTPClient_GetLastActivityQuestion_InvalidJSON(t *testing.T) {
	// Arrange: сервер возвращает некорректный JSON.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer ts.Close()

	client := newTestClient(ts.URL, 5*time.Second)
	link := "https://stackoverflow.com/questions/12345/some-question-title"

	// Act
	_, err := client.GetLastActivityQuestion(link)

	// Assert: ошибка декодирования JSON.
	require.Error(t, err)
}

func TestStackOverflowHTTPClient_GetLastActivityQuestion_InvalidLink(t *testing.T) {
	// Arrange: передаём ссылку, которая не соответствует ожидаемому формату.
	client := clients.NewStackOverflowHTTPClient()
	invalidLink := "https://stackoverflow.com/users/12345" // не questions

	// Act
	_, err := client.GetLastActivityQuestion(invalidLink)

	// Assert: ожидается ошибка, сформированная в apiSOUrlGeneration.
	require.Error(t, err)
	assert.Contains(t, err.Error(), "wrong url format")
}
