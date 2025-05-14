package links_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"LinkTracker/internal/domain"
	"LinkTracker/internal/infrastructure/dto"

	scrapperdto "LinkTracker/internal/infrastructure/dto/dto_scrapper"
	"LinkTracker/internal/infrastructure/httpapi/links"
	"LinkTracker/internal/infrastructure/httpapi/links/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_DeleteLinksHandler_ServeHTTP_Success(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	link := domain.Link{URL: "https://example.com/example"}
	linkDTO := dto.LinkToRemoveListRequestDTO(&link)

	payload, err := json.Marshal(linkDTO)
	require.NoError(t, err)

	linkDeleter := &mocks.LinkDeleter{}
	linkDeleter.On("DeleteLink", ctx, tgID, &link).Return(link, nil)

	deleteLinksHandler := links.DeleteLinksHandler{LinkDeleter: linkDeleter}

	r := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/links", bytes.NewReader(payload))
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	deleteLinksHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_DeleteLinksHandler_ServeHTTP_InvalidTgID(t *testing.T) {
	ctx := context.Background()
	tgID := "invalidTgID"
	linkDeleter := &mocks.LinkDeleter{}
	deleteLinksHandler := links.DeleteLinksHandler{LinkDeleter: linkDeleter}

	r := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/links", http.NoBody)
	r.Header.Set("Tg-Chat-Id", tgID)

	w := httptest.NewRecorder()

	deleteLinksHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Invalid or missing tgID", *responseErrorBody.Description)
	assert.Equal(t, "400", *responseErrorBody.Code)
	assert.Equal(t, "INVALID_TG_ID", *responseErrorBody.ExceptionName)
}

func Test_DeleteLinksHandler_ServeHTTP_InvalidBody(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	linkDeleter := &mocks.LinkDeleter{}
	deleteLinksHandler := links.DeleteLinksHandler{LinkDeleter: linkDeleter}

	testcases := []struct {
		testName string
		payload  io.Reader
	}{
		{testName: "payload=http.NoBody", payload: http.NoBody},
		{testName: "payload=nil", payload: nil},
		{testName: "payload=invalidBody", payload: bytes.NewReader([]byte("invalidBody"))},
		{testName: "payload=\"\"", payload: bytes.NewReader([]byte(""))},
	}

	for _, tc := range testcases {
		t.Run(tc.testName, func(t *testing.T) {
			r := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/links", tc.payload)
			r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

			w := httptest.NewRecorder()

			deleteLinksHandler.ServeHTTP(w, r)

			var responseErrorBody scrapperdto.ApiErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Equal(t, "Invalid or missing request body", *responseErrorBody.Description)
			assert.Equal(t, "400", *responseErrorBody.Code)
			assert.Equal(t, "INVALID_REQUEST_BODY", *responseErrorBody.ExceptionName)
		})
	}
}

func Test_DeleteLinksHandler_ServeHTTP_InvalidBody_MissingRequiredFields(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)

	linkDeleter := &mocks.LinkDeleter{}
	deleteLinksHandler := links.DeleteLinksHandler{LinkDeleter: linkDeleter}

	payload := bytes.NewReader([]byte("{}"))

	r := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/links", payload)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

	w := httptest.NewRecorder()

	deleteLinksHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Missing required fields", *responseErrorBody.Description)
	assert.Equal(t, "400", *responseErrorBody.Code)
	assert.Equal(t, "INVALID_REQUEST_BODY", *responseErrorBody.ExceptionName)
}

func Test_DeleteLinksHandler_ServeHTTP_DeleteLinkError_LinkNotExist(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	link := domain.Link{URL: "https://example.com/example"}
	linkDTO := dto.LinkToRemoveListRequestDTO(&link)

	payload, err := json.Marshal(linkDTO)
	require.NoError(t, err)

	linkDeleter := &mocks.LinkDeleter{}
	linkDeleter.On("DeleteLink", ctx, tgID, &link).Return(domain.Link{}, domain.ErrLinkNotExist{})
	deleteLinksHandler := links.DeleteLinksHandler{LinkDeleter: linkDeleter}

	r := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/links", bytes.NewReader(payload))
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	deleteLinksHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "Link not found", *responseErrorBody.Description)
	assert.Equal(t, "404", *responseErrorBody.Code)
	assert.Equal(t, "LINK_NOT_EXIST", *responseErrorBody.ExceptionName)
}

func Test_DeleteLinksHandler_ServeHTTP_DeleteLinkError(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	link := domain.Link{URL: "https://example.com/example"}
	linkDTO := dto.LinkToRemoveListRequestDTO(&link)
	payload, err := json.Marshal(linkDTO)
	require.NoError(t, err)

	linkDeleter := &mocks.LinkDeleter{}
	linkDeleter.On("DeleteLink", ctx, tgID, &link).Return(domain.Link{}, errors.New("some error"))
	deleteLinksHandler := links.DeleteLinksHandler{LinkDeleter: linkDeleter}

	r := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/links", bytes.NewReader(payload))
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	deleteLinksHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Failed to delete link", *responseErrorBody.Description)
	assert.Equal(t, "500", *responseErrorBody.Code)
	assert.Equal(t, "DELETE_LINK_FAILED", *responseErrorBody.ExceptionName)
}
