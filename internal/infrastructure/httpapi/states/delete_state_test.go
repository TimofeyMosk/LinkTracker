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

	scrapperdto "LinkTracker/internal/infrastructure/dto/dto_scrapper"
	"LinkTracker/internal/infrastructure/httpapi/states"
	"LinkTracker/internal/infrastructure/httpapi/states/mocks"
)

func Test_DeleteStatesHandler_ServeHTTP_InvalidTgId(t *testing.T) {
	ctx := context.Background()
	tgID := "invalidTgID"

	stateDeleter := &mocks.StateDeleter{}
	deleteStatesHandler := states.DeleteStatesHandler{StateDeleter: stateDeleter}

	r := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/states", http.NoBody)
	r.Header.Set("Tg-Chat-Id", tgID)

	w := httptest.NewRecorder()

	deleteStatesHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Invalid or missing tgID", *responseErrorBody.Description)
	assert.Equal(t, "400", *responseErrorBody.Code)
	assert.Equal(t, "INVALID_TG_ID", *responseErrorBody.ExceptionName)
}

func Test_DeleteStatesHandler_ServeHTTP_DeleteStateError(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)

	stateDeleter := &mocks.StateDeleter{}
	stateDeleter.On("DeleteState", ctx, tgID).Return(errors.New("some error"))
	deleteStatesHandler := states.DeleteStatesHandler{StateDeleter: stateDeleter}

	r := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/states", http.NoBody)
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

	w := httptest.NewRecorder()

	deleteStatesHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Failed to delete state", *responseErrorBody.Description)
	assert.Equal(t, "500", *responseErrorBody.Code)
	assert.Equal(t, "DELETE_STATE_FAILED", *responseErrorBody.ExceptionName)
}

func Test_DeleteStatesHandler_ServeHTTP_Success(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)

	stateDeleter := &mocks.StateDeleter{}
	stateDeleter.On("DeleteState", ctx, tgID).Return(nil)
	deleteStatesHandler := states.DeleteStatesHandler{StateDeleter: stateDeleter}

	r := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/states", http.NoBody)
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

	w := httptest.NewRecorder()

	deleteStatesHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}
