package states_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	scrapperdto "LinkTracker/internal/infrastructure/dto/dto_scrapper"
	"LinkTracker/internal/infrastructure/httpapi/states"
	"LinkTracker/internal/infrastructure/httpapi/states/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_PostStatesHandler_ServeHTTP_InvalidTgId(t *testing.T) {
	ctx := context.Background()
	tgID := "invalidTgID"

	stateCreator := &mocks.StateCreator{}
	postStatesHandler := states.PostStatesHandler{StateCreator: stateCreator}

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/states", http.NoBody)
	r.Header.Set("Tg-Chat-Id", tgID)

	w := httptest.NewRecorder()

	postStatesHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Invalid or missing tgID", *responseErrorBody.Description)
	assert.Equal(t, "400", *responseErrorBody.Code)
	assert.Equal(t, "INVALID_TG_ID", *responseErrorBody.ExceptionName)
}

func Test_PostStatesHandler_ServeHTTP_InvalidBody(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	payload := []byte("invalidJson")

	stateCreator := &mocks.StateCreator{}
	postStatesHandler := states.PostStatesHandler{StateCreator: stateCreator}

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/states", bytes.NewBuffer(payload))
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

	w := httptest.NewRecorder()

	postStatesHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Invalid or missing request body", *responseErrorBody.Description)
	assert.Equal(t, "400", *responseErrorBody.Code)
	assert.Equal(t, "INVALID_REQUEST_BODY", *responseErrorBody.ExceptionName)
}

func Test_PostStatesHandler_ServeHTTP_InvalidBody_MissingRequiredFields(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	payload := []byte("{}")

	stateCreator := &mocks.StateCreator{}
	postStatesHandler := states.PostStatesHandler{StateCreator: stateCreator}

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/states", bytes.NewBuffer(payload))
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

	w := httptest.NewRecorder()

	postStatesHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Missing required fields", *responseErrorBody.Description)
	assert.Equal(t, "400", *responseErrorBody.Code)
	assert.Equal(t, "INVALID_REQUEST_BODY", *responseErrorBody.ExceptionName)
}

func Test_PostStatesHandler_ServeHTTP_CreateStateError(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	state := 1
	stateRequest := scrapperdto.StateRequest{State: &state}
	payload, err := json.Marshal(stateRequest)
	require.NoError(t, err)

	stateCreator := &mocks.StateCreator{}
	stateCreator.On("CreateState", ctx, tgID, state).Return(errors.New("some error"))
	postStatesHandler := states.PostStatesHandler{StateCreator: stateCreator}

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/states", bytes.NewBuffer(payload))
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

	w := httptest.NewRecorder()

	postStatesHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Failed to create state", *responseErrorBody.Description)
	assert.Equal(t, "500", *responseErrorBody.Code)
	assert.Equal(t, "CREATE_STATE_FAILED", *responseErrorBody.ExceptionName)
}

func Test_PostStatesHandler_ServeHTTP_Success(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	state := 1
	stateRequest := scrapperdto.StateRequest{State: &state}
	payload, err := json.Marshal(stateRequest)
	require.NoError(t, err)

	stateCreator := &mocks.StateCreator{}
	stateCreator.On("CreateState", ctx, tgID, state).Return(nil)
	postStatesHandler := states.PostStatesHandler{StateCreator: stateCreator}

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/states", bytes.NewBuffer(payload))
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

	w := httptest.NewRecorder()

	postStatesHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}
