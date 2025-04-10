package tgchat_test

//func Test_DeleteUserHandler_InvalidChatID(t *testing.T) {
//	mockScrapper := &mocks.UserDeleter{}
//	handler := tgchat.DeleteUserHandler{UserDeleter: mockScrapper}
//	id := "errorID"
//	request := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s", id), http.NoBody)
//	request.SetPathValue("id", id)
//
//	responseRecorder := httptest.NewRecorder()
//
//	handler.ServeHTTP(responseRecorder, request)
//
//	var apiErrorBody scrapperdto.ApiErrorResponse
//	err := json.Unmarshal(responseRecorder.Body.Bytes(), &apiErrorBody)
//	assert.NoError(t, err)
//	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
//	assert.Equal(t, "INVALID_CHAT_ID", *apiErrorBody.Code)
//	assert.Equal(t, "Invalid or missing chat ID", *apiErrorBody.Description)
//	assert.Equal(t, "BadRequest", *apiErrorBody.ExceptionName)
//}
//
//func Test_DeleteUserHandler_UserNotExist(t *testing.T) {
//	mockScrapper := &mocks.UserDeleter{}
//	tgID := int64(123)
//	mockScrapper.On("DeleteUser", tgID).Return(domain.ErrUserNotExist{}).Once()
//
//	handler := tgchat.DeleteUserHandler{UserDeleter: mockScrapper}
//	request := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%d", tgID), http.NoBody)
//	request.SetPathValue("id", strconv.Itoa(int(tgID)))
//
//	responseRecorder := httptest.NewRecorder()
//
//	handler.ServeHTTP(responseRecorder, request)
//
//	var apiErrorBody scrapperdto.ApiErrorResponse
//	err := json.Unmarshal(responseRecorder.Body.Bytes(), &apiErrorBody)
//	assert.NoError(t, err)
//	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)
//	assert.Equal(t, "CHAT_NOT_EXIST", *apiErrorBody.Code)
//	assert.Equal(t, "Chat not exist", *apiErrorBody.Description)
//	assert.Equal(t, "Not Found", *apiErrorBody.ExceptionName)
//	mockScrapper.AssertExpectations(t)
//}
//
//func Test_DeleteUserHandler_ChatNotDeleted(t *testing.T) {
//	mockScrapper := &mocks.UserDeleter{}
//	tgID := int64(123)
//	mockScrapper.On("DeleteUser", tgID).Return(errors.New("some error")).Once()
//
//	handler := tgchat.DeleteUserHandler{UserDeleter: mockScrapper}
//	request := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%d", tgID), http.NoBody)
//	request.SetPathValue("id", strconv.Itoa(int(tgID)))
//
//	responseRecorder := httptest.NewRecorder()
//
//	handler.ServeHTTP(responseRecorder, request)
//
//	var apiErrorBody scrapperdto.ApiErrorResponse
//	err := json.Unmarshal(responseRecorder.Body.Bytes(), &apiErrorBody)
//	assert.NoError(t, err)
//	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
//	assert.Equal(t, "CHAT_NOT_DELETED", *apiErrorBody.Code)
//	assert.Equal(t, "Chat has not been deleted", *apiErrorBody.Description)
//	assert.Equal(t, "BadRequest", *apiErrorBody.ExceptionName)
//	mockScrapper.AssertExpectations(t)
//}
//
//func Test_DeleteUserHandler_Success(t *testing.T) {
//	mockScrapper := &mocks.UserDeleter{}
//	tgID := int64(123)
//	mockScrapper.On("DeleteUser", tgID).Return(nil).Once()
//
//	handler := tgchat.DeleteUserHandler{UserDeleter: mockScrapper}
//	request := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%d", tgID), http.NoBody)
//	request.SetPathValue("id", strconv.Itoa(int(tgID)))
//
//	responseRecorder := httptest.NewRecorder()
//
//	handler.ServeHTTP(responseRecorder, request)
//
//	assert.Equal(t, http.StatusOK, responseRecorder.Code)
//	mockScrapper.AssertExpectations(t)
//}
