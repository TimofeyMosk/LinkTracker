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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	scrapperdto "LinkTracker/internal/infrastructure/dto/dto_scrapper"
	"LinkTracker/internal/infrastructure/httpapi/links"
	"LinkTracker/internal/infrastructure/httpapi/links/mocks"
)

func TestPostLinksHandler_ServeHTTP_Succes(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	linkID := int64(100)
	link := domain.Link{URL: "https://example.com/example",
		Tags: []string{"tag", "tag2"}, Filters: []string{"filter1, filter2"}}
	responseLink := domain.Link{URL: "https://example.com/example",
		Tags: []string{"tag", "tag2"}, Filters: []string{"filter1, filter2"}, ID: linkID}
	linkDTO := dto.LinkToLinkRequestDTO(&link)
	payload, err := json.Marshal(linkDTO)
	require.NoError(t, err)

	linkAdder := &mocks.LinkAdder{}
	linkAdder.On("AddLink", ctx, tgID, &link).Return(responseLink, nil)
	postLinksHandler := links.PostLinksHandler{LinkAdder: linkAdder}

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/links", bytes.NewReader(payload))
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))
	r.Header.Set("content-type", "application/json")

	w := httptest.NewRecorder()

	postLinksHandler.ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)

	var linkRequest scrapperdto.LinkResponse
	err = json.Unmarshal(w.Body.Bytes(), &linkRequest)
	require.NoError(t, err)
	assert.Equal(t, link.URL, *linkRequest.Url)
	assert.Equal(t, link.Tags, *linkRequest.Tags)
	assert.Equal(t, link.Filters, *linkRequest.Filters)
	assert.Equal(t, linkID, *linkRequest.Id)
}

func Test_PostLinkHandler_ServeHTTP_InvalidTgID(t *testing.T) {
	ctx := context.Background()
	tgID := "invalidTgID"
	linkAdder := &mocks.LinkAdder{}
	postLinksHandler := links.PostLinksHandler{LinkAdder: linkAdder}

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/links", http.NoBody)
	r.Header.Set("Tg-Chat-Id", tgID)

	w := httptest.NewRecorder()

	postLinksHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Invalid or missing tgID", *responseErrorBody.Description)
	assert.Equal(t, "400", *responseErrorBody.Code)
	assert.Equal(t, "INVALID_TG_ID", *responseErrorBody.ExceptionName)
}

func Test_PostLinkHandler_ServeHTTP_InvalidBody(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	linkAdder := &mocks.LinkAdder{}
	postLinksHandler := links.PostLinksHandler{LinkAdder: linkAdder}

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
			r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/links", tc.payload)
			r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

			w := httptest.NewRecorder()

			postLinksHandler.ServeHTTP(w, r)

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

func Test_PostLinkHandler_ServeHTTP_InvalidBody_MissingRequiredFields(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)

	linkAdder := &mocks.LinkAdder{}
	postLinksHandler := links.PostLinksHandler{LinkAdder: linkAdder}

	payload := bytes.NewReader([]byte("{}"))

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/links", payload)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

	w := httptest.NewRecorder()

	postLinksHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Missing required fields", *responseErrorBody.Description)
	assert.Equal(t, "400", *responseErrorBody.Code)
	assert.Equal(t, "INVALID_REQUEST_BODY", *responseErrorBody.ExceptionName)
}

func Test_PostLinkHandler_ServeHTTP_AddLinkError_UserNotExist(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	link := domain.Link{URL: "https://example.com/example",
		Tags: []string{"tag", "tag2"}, Filters: []string{"filter1, filter2"}}
	linkDTO := dto.LinkToLinkRequestDTO(&link)
	payload, err := json.Marshal(linkDTO)
	require.NoError(t, err)

	linkAdder := &mocks.LinkAdder{}
	linkAdder.On("AddLink", ctx, tgID, &link).Return(domain.Link{}, domain.ErrUserNotExist{})
	postLinksHandler := links.PostLinksHandler{LinkAdder: linkAdder}

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/links", bytes.NewReader(payload))
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	postLinksHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "User not exist", *responseErrorBody.Description)
	assert.Equal(t, "404", *responseErrorBody.Code)
	assert.Equal(t, "USER_NOT_EXIST", *responseErrorBody.ExceptionName)
}

func Test_PostLinkHandler_ServeHTTP_AddLinkError_ErrLinkAlreadyTracking(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	link := domain.Link{URL: "https://example.com/example",
		Tags: []string{"tag", "tag2"}, Filters: []string{"filter1, filter2"}}
	linkDTO := dto.LinkToLinkRequestDTO(&link)
	payload, err := json.Marshal(linkDTO)
	require.NoError(t, err)

	linkAdder := &mocks.LinkAdder{}
	linkAdder.On("AddLink", ctx, tgID, &link).Return(domain.Link{}, domain.ErrLinkAlreadyTracking{})
	postLinksHandler := links.PostLinksHandler{LinkAdder: linkAdder}

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/links", bytes.NewReader(payload))
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	postLinksHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Links already tracking", *responseErrorBody.Description)
	assert.Equal(t, "400", *responseErrorBody.Code)
	assert.Equal(t, "LINKS_ALREADY_EXIST", *responseErrorBody.ExceptionName)
}

func Test_PostLinkHandler_ServeHTTP_AddLinkError(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	link := domain.Link{URL: "https://example.com/example",
		Tags: []string{"tag", "tag2"}, Filters: []string{"filter1, filter2"}}
	linkDTO := dto.LinkToLinkRequestDTO(&link)
	payload, err := json.Marshal(linkDTO)
	require.NoError(t, err)

	linkAdder := &mocks.LinkAdder{}
	linkAdder.On("AddLink", ctx, tgID, &link).Return(domain.Link{}, errors.New("some error"))
	postLinksHandler := links.PostLinksHandler{LinkAdder: linkAdder}

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/links", bytes.NewReader(payload))
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	postLinksHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Failed to added link", *responseErrorBody.Description)
	assert.Equal(t, "500", *responseErrorBody.Code)
	assert.Equal(t, "ADD_LINK_FAILED", *responseErrorBody.ExceptionName)
}
