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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"LinkTracker/internal/domain"
	scrapperdto "LinkTracker/internal/infrastructure/dto/dto_scrapper"
	"LinkTracker/internal/infrastructure/httpapi/states"
	"LinkTracker/internal/infrastructure/httpapi/states/mocks"
)

func Test_PutStatesHandler_ServeHTTP_InvalidTgId(t *testing.T) {
	ctx := context.Background()
	tgID := "invalidTgID"

	stateUpdater := &mocks.StateUpdater{}
	putStatesHandler := states.PutStatesHandler{StateUpdater: stateUpdater}

	r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/states", http.NoBody)
	r.Header.Set("Tg-Chat-Id", tgID)

	w := httptest.NewRecorder()

	putStatesHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Invalid or missing tgID", *responseErrorBody.Description)
	assert.Equal(t, "400", *responseErrorBody.Code)
	assert.Equal(t, "INVALID_TG_ID", *responseErrorBody.ExceptionName)
}

func Test_PutStatesHandler_ServeHTTP_InvalidBody(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	payload := []byte("invalidJson")

	stateUpdater := &mocks.StateUpdater{}
	putStatesHandler := states.PutStatesHandler{StateUpdater: stateUpdater}

	r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/states", bytes.NewReader(payload))
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))
	r.Header.Set("content-type", "application/json")

	w := httptest.NewRecorder()

	putStatesHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Invalid or missing request body", *responseErrorBody.Description)
	assert.Equal(t, "400", *responseErrorBody.Code)
	assert.Equal(t, "INVALID_REQUEST_BODY", *responseErrorBody.ExceptionName)
}

func Test_PutStatesHandler_ServeHTTP_UpdateStateError(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	state := 1
	link := "https://example.com/example"
	tags := []string{"tag1", "tag2"}
	filters := []string{"filter1", "filter2"}
	stateRequest := scrapperdto.StateRequest{State: &state, Link: &link, Tags: &tags, Filters: &filters}

	payload, err := json.Marshal(stateRequest)
	require.NoError(t, err)

	stateUpdater := &mocks.StateUpdater{}
	stateUpdater.On("UpdateState", ctx, tgID, state, &domain.Link{URL: link, Tags: tags, Filters: filters}).
		Return(errors.New("some error"))

	putStatesHandler := states.PutStatesHandler{StateUpdater: stateUpdater}

	r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/states", bytes.NewReader(payload))
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))
	r.Header.Set("content-type", "application/json")

	w := httptest.NewRecorder()

	putStatesHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Failed to update state", *responseErrorBody.Description)
	assert.Equal(t, "500", *responseErrorBody.Code)
	assert.Equal(t, "UPDATE_STATE_ERROR", *responseErrorBody.ExceptionName)
}

func Test_PutStatesHandler_ServeHTTP_Success(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	state := 1
	link := "https://example.com/example"
	tags := []string{"tag1", "tag2"}
	filters := []string{"filter1", "filter2"}
	stateRequest := scrapperdto.StateRequest{State: &state, Link: &link, Tags: &tags, Filters: &filters}

	payload, err := json.Marshal(stateRequest)
	require.NoError(t, err)

	stateUpdater := &mocks.StateUpdater{}
	stateUpdater.On("UpdateState", ctx, tgID, state, &domain.Link{URL: link, Tags: tags, Filters: filters}).Return(nil)
	putStatesHandler := states.PutStatesHandler{StateUpdater: stateUpdater}

	r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/states", bytes.NewReader(payload))
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))
	r.Header.Set("content-type", "application/json")

	w := httptest.NewRecorder()

	putStatesHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}
