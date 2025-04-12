package updates_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	botdto "LinkTracker/internal/infrastructure/dto/dto_bot"
	"LinkTracker/internal/infrastructure/httpapi/updates"
	"LinkTracker/internal/infrastructure/httpapi/updates/mocks"

	"github.com/stretchr/testify/assert"
)

func Test_PostUpdatesHandler_Success(t *testing.T) {
	ctx := context.Background()
	tgIDs := []int64{12345, 67890}
	url := "https://example.com/update"
	description := "Было обновление : https://example.com/update"
	requestBody := botdto.LinkUpdate{
		TgChatIds:   &tgIDs,
		Url:         &url,
		Description: &description,
	}
	payload, _ := json.Marshal(requestBody)

	bot := &mocks.UpdateSender{}
	handler := updates.PostUpdatesHandler{UpdateSender: bot}

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/updates", bytes.NewReader(payload))
	w := httptest.NewRecorder()

	bot.On("UpdateSend", ctx, tgIDs, url, description)

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	bot.AssertExpectations(t)
}

func Test_PostUpdatesHandler_InvalidBody(t *testing.T) {
	updateSender := &mocks.UpdateSender{}
	postUpdatesHandler := updates.PostUpdatesHandler{UpdateSender: updateSender}
	payload := []byte("invalid json")

	request := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(payload))
	responseRecorder := httptest.NewRecorder()

	postUpdatesHandler.ServeHTTP(responseRecorder, request)

	var responseData botdto.ApiErrorResponse
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseData)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	assert.Equal(t, "Invalid or missing request body", *responseData.Description)
	assert.Equal(t, "400", *responseData.Code)
	assert.Equal(t, "INVALID_REQUEST_BODY", *responseData.ExceptionName)
}

func Test_PostUpdatesHandler_MissingRequiredFields(t *testing.T) {
	updateSender := &mocks.UpdateSender{}
	postUpdatesHandler := updates.PostUpdatesHandler{UpdateSender: updateSender}
	emptyLinkUpdateDTO := botdto.LinkUpdate{}
	payload, _ := json.Marshal(emptyLinkUpdateDTO)

	r := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(payload))
	w := httptest.NewRecorder()

	postUpdatesHandler.ServeHTTP(w, r)

	var responseData botdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &responseData)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "\"TgChatIds\" or \"Url\" is missing", *responseData.Description)
	assert.Equal(t, "400", *responseData.Code)
	assert.Equal(t, "MISSING_REQUIRED_FIELDS", *responseData.ExceptionName)
}
