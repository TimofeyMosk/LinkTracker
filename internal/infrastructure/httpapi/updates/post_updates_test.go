package updates_test

import (
	"bytes"
	"encoding/json"

	botdto "LinkTracker/internal/infrastructure/dto/dto_bot"
	"LinkTracker/internal/infrastructure/httpapi/updates"
	"LinkTracker/internal/infrastructure/httpapi/updates/mocks"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_PostUpdatesHandler_InvalidRequestBody(t *testing.T) {
	bot := &mocks.UpdateSender{}
	handler := updates.PostUpdatesHandler{UpdateSender: bot}
	request := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader([]byte("invalid json")))
	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	var responseData botdto.ApiErrorResponse
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseData)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	assert.Equal(t, "INVALID_REQUEST_BODY", *responseData.Code)
	assert.Equal(t, "Invalid or missing request body", *responseData.Description)
	assert.Equal(t, "BadRequest", *responseData.ExceptionName)
}

func Test_PostUpdatesHandler_ValidRequest(t *testing.T) {
	bot := &mocks.UpdateSender{}
	handler := updates.PostUpdatesHandler{UpdateSender: bot}

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

func Test_PostUpdatesHandler_MissingRequiredFields(t *testing.T) {
	bot := &mocks.UpdateSender{}
	handler := updates.PostUpdatesHandler{UpdateSender: bot}
	requestBody := botdto.LinkUpdate{}
	body, _ := json.Marshal(requestBody)

	request := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	var responseData botdto.ApiErrorResponse
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseData)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	assert.Equal(t, "MISSING_REQUIRED_FIELDS", *responseData.Code)
	assert.Equal(t, "\"TgChatIds\" or \"Url\" is missing", *responseData.Description)
	assert.Equal(t, "BadRequest", *responseData.ExceptionName)
}
