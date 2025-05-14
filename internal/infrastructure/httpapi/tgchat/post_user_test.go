package tgchat_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	scrapperdto "LinkTracker/internal/infrastructure/dto/dto_scrapper"
	"LinkTracker/internal/infrastructure/httpapi/tgchat"
	"LinkTracker/internal/infrastructure/httpapi/tgchat/mocks"
)

func Test_PostUserHandler_ServeHTTP_InvalidChatID(t *testing.T) {
	ctx := context.Background()
	tgID := "invalidID"
	userAdder := &mocks.UserAdder{}
	handler := tgchat.PostUserHandler{UserAdder: userAdder}

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/users/%s", tgID), http.NoBody)
	r.SetPathValue("id", tgID)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	var apiErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &apiErrorBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Invalid or missing tgID", *apiErrorBody.Description)
	assert.Equal(t, "400", *apiErrorBody.Code)
	assert.Equal(t, "INVALID_TG_ID", *apiErrorBody.ExceptionName)
}

func Test_PostUserHandler_ServeHTTP_AddUserError(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	userAdder := &mocks.UserAdder{}
	userAdder.On("AddUser", ctx, tgID).Return(errors.New("some error"))
	handler := tgchat.PostUserHandler{UserAdder: userAdder}

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/users/%d", tgID), http.NoBody)
	r.SetPathValue("id", strconv.FormatInt(tgID, 10))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	var apiErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &apiErrorBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Failed to create user", *apiErrorBody.Description)
	assert.Equal(t, "500", *apiErrorBody.Code)
	assert.Equal(t, "CREATE_USER_FAILED", *apiErrorBody.ExceptionName)
}

func Test_PostUserHandler_ServeHTTP_Success(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	userAdder := &mocks.UserAdder{}
	userAdder.On("AddUser", ctx, tgID).Return(nil)
	handler := tgchat.PostUserHandler{UserAdder: userAdder}

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/users/%d", tgID), http.NoBody)
	r.SetPathValue("id", strconv.FormatInt(tgID, 10))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}
