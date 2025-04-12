package states_test

import (
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

func Test_GetStatesHandler_ServeHTTP_InvalidTgId(t *testing.T) {
	ctx := context.Background()
	tgID := "invalidTgID"

	stateGetter := &mocks.StateGetter{}
	getStatesHandler := states.GetStatesHandler{StateGetter: stateGetter}

	r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/states", http.NoBody)
	r.Header.Set("Tg-Chat-Id", tgID)

	w := httptest.NewRecorder()

	getStatesHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Invalid or missing tgID", *responseErrorBody.Description)
	assert.Equal(t, "400", *responseErrorBody.Code)
	assert.Equal(t, "INVALID_TG_ID", *responseErrorBody.ExceptionName)
}

func Test_GetStatesHandler_ServeHTTP_GetStateError(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)

	stateGetter := &mocks.StateGetter{}
	stateGetter.On("GetState", ctx, tgID).Return(0, domain.Link{}, errors.New("some error"))
	getStatesHandler := states.GetStatesHandler{StateGetter: stateGetter}

	r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/states", http.NoBody)
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

	w := httptest.NewRecorder()

	getStatesHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Failed to get state", *responseErrorBody.Description)
	assert.Equal(t, "500", *responseErrorBody.Code)
	assert.Equal(t, "GET_STATE_FAILED", *responseErrorBody.ExceptionName)
}

func Test_GetStatesHandler_ServeHTTP_Success(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	state := 1
	link := "https://example.com/example"
	tags := []string{"tag1", "tag2"}
	filters := []string{"filter1", "filter2"}
	stateGetter := &mocks.StateGetter{}
	stateGetter.On("GetState", ctx, tgID).Return(state, domain.Link{URL: link, Tags: tags, Filters: filters}, nil)
	getStatesHandler := states.GetStatesHandler{StateGetter: stateGetter}

	r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/states", http.NoBody)
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

	w := httptest.NewRecorder()

	getStatesHandler.ServeHTTP(w, r)

	var stateResponse scrapperdto.StateResponse
	err := json.Unmarshal(w.Body.Bytes(), &stateResponse)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, state, *stateResponse.State)
	assert.Equal(t, link, *stateResponse.Link)
	assert.Equal(t, tags, *stateResponse.Tags)
	assert.Equal(t, filters, *stateResponse.Filters)
}
