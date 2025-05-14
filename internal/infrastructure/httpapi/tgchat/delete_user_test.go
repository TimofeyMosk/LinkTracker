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

	"LinkTracker/internal/domain"

	"github.com/stretchr/testify/assert"

	scrapperdto "LinkTracker/internal/infrastructure/dto/dto_scrapper"
	"LinkTracker/internal/infrastructure/httpapi/tgchat"
	"LinkTracker/internal/infrastructure/httpapi/tgchat/mocks"
)

func Test_DeleteUserHandler_InvalidChatID(t *testing.T) {
	ctx := context.Background()
	tgID := "invalidID"
	userDeleter := &mocks.UserDeleter{}
	handler := tgchat.DeleteUserHandler{UserDeleter: userDeleter}

	r := httptest.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("/users/%s", tgID), http.NoBody)
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

func Test_DeleteUserHandler_UserNotExist(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	userDeleter := &mocks.UserDeleter{}
	userDeleter.On("DeleteUser", ctx, tgID).Return(domain.ErrUserNotExist{})
	handler := tgchat.DeleteUserHandler{UserDeleter: userDeleter}

	r := httptest.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("/users/%d", tgID), http.NoBody)
	r.SetPathValue("id", strconv.FormatInt(tgID, 10))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	var apiErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &apiErrorBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "User not exist", *apiErrorBody.Description)
	assert.Equal(t, "404", *apiErrorBody.Code)
	assert.Equal(t, "USER_NOT_EXIST", *apiErrorBody.ExceptionName)
}

func Test_DeleteUserHandler_ChatNotDeleted(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	userDeleter := &mocks.UserDeleter{}
	userDeleter.On("DeleteUser", ctx, tgID).Return(errors.New("some error"))
	handler := tgchat.DeleteUserHandler{UserDeleter: userDeleter}

	r := httptest.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("/users/%d", tgID), http.NoBody)
	r.SetPathValue("id", strconv.FormatInt(tgID, 10))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	var apiErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &apiErrorBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Chat has not been deleted", *apiErrorBody.Description)
	assert.Equal(t, "500", *apiErrorBody.Code)
	assert.Equal(t, "CHAT_NOT_DELETED", *apiErrorBody.ExceptionName)
}

func Test_DeleteUserHandler_Success(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	userDeleter := &mocks.UserDeleter{}
	userDeleter.On("DeleteUser", ctx, tgID).Return(nil)
	handler := tgchat.DeleteUserHandler{UserDeleter: userDeleter}

	r := httptest.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("/users/%d", tgID), http.NoBody)
	r.SetPathValue("id", strconv.FormatInt(tgID, 10))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}
