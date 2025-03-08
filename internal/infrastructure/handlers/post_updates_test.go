package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	botdto "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/dto/dto_bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/handlers"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/handlers/mocks"
)

func TestPostUpdatesHandler_InvalidRequestBody(t *testing.T) {
	bot := &mocks.Bot{}
	handler := handlers.PostUpdatesHandler{Bot: bot}
	request := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader([]byte("invalid json")))
	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
}

func TestPostUpdatesHandler_ValidRequest(t *testing.T) {
	bot := &mocks.Bot{}
	handler := handlers.PostUpdatesHandler{Bot: bot}

	chatIDs := []int64{12345, 67890}
	url := "https://example.com/update"
	requestBody := botdto.LinkUpdate{
		TgChatIds: &chatIDs,
		Url:       &url,
	}
	body, _ := json.Marshal(requestBody)

	request := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
	responseRecorder := httptest.NewRecorder()

	bot.On("SendMessage", int64(12345), "Было обновление : https://example.com/update").Return()
	bot.On("SendMessage", int64(67890), "Было обновление : https://example.com/update").Return()

	handler.ServeHTTP(responseRecorder, request)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	bot.AssertExpectations(t)
}

func TestPostUpdatesHandler_MissingRequiredFields(t *testing.T) {
	bot := &mocks.Bot{}
	handler := handlers.PostUpdatesHandler{Bot: bot}
	requestBody := botdto.LinkUpdate{}
	body, _ := json.Marshal(requestBody)

	request := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
}
