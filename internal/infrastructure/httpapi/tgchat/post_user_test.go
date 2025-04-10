package tgchat_test

//func Test_PostUserHandler_Success(t *testing.T) {
//	mockScrapper := &mocks.UserAdder{}
//	mockScrapper.On("AddUser", mock.Anything).Return(nil).Once()
//
//	handler := tgchat.PostUserHandler{UserAdder: mockScrapper}
//
//	request := httptest.NewRequest(http.MethodPost, "/users/123", http.NoBody)
//	request.SetPathValue("id", "123")
//
//	responseRecorder := httptest.NewRecorder()
//	handler.ServeHTTP(responseRecorder, request)
//
//	assert.Equal(t, http.StatusOK, responseRecorder.Code)
//	assert.Empty(t, responseRecorder.Body.String())
//
//	mockScrapper.AssertExpectations(t)
//}
//
//func Test_PostUserHandler_InvalidChatID(t *testing.T) {
//	mockScrapper := &mocks.UserAdder{}
//	handler := tgchat.PostUserHandler{UserAdder: mockScrapper}
//	request := httptest.NewRequest(http.MethodPost, "/users/invalid", http.NoBody)
//	request.SetPathValue("id", "invalid")
//
//	responseRecorder := httptest.NewRecorder()
//	handler.ServeHTTP(responseRecorder, request)
//
//	var apiErrorBody scrapperdto.ApiErrorResponse
//	err := json.Unmarshal(responseRecorder.Body.Bytes(), &apiErrorBody)
//	assert.NoError(t, err)
//
//	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
//	assert.Equal(t, "INVALID_CHAT_ID", *apiErrorBody.Code)
//	assert.Equal(t, "Invalid or missing chat ID", *apiErrorBody.Description)
//	assert.Equal(t, "BadRequest", *apiErrorBody.ExceptionName)
//}
//
//func Test_PostUserHandler_ScrapperError(t *testing.T) {
//	mockScrapper := &mocks.UserAdder{}
//	mockScrapper.On("AddUser", mock.Anything).Return(errors.New("some error")).Once()
//
//	handler := tgchat.PostUserHandler{UserAdder: mockScrapper}
//
//	request := httptest.NewRequest(http.MethodPost, "/users/123", http.NoBody)
//	request.SetPathValue("id", "123")
//
//	responseRecorder := httptest.NewRecorder()
//	handler.ServeHTTP(responseRecorder, request)
//
//	var apiErrorBody scrapperdto.ApiErrorResponse
//	err := json.Unmarshal(responseRecorder.Body.Bytes(), &apiErrorBody)
//	assert.NoError(t, err)
//
//	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
//	assert.Equal(t, "CREATE_CHAT_FAILED", *apiErrorBody.Code)
//	assert.Equal(t, "Failed to create chat", *apiErrorBody.Description)
//	assert.Equal(t, "BadRequest", *apiErrorBody.ExceptionName)
//
//	mockScrapper.AssertExpectations(t)
//}
