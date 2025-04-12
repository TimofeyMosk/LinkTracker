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

func Test_PutLinksHandler_ServeHTTP_Success(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	testCases := []struct {
		testName string
		link     domain.Link
	}{
		{
			testName: "exampleURLWithoutTagsAndFilters",
			link:     domain.Link{URL: "https://example.com/example", Tags: []string{}, Filters: []string{}},
		},
		{
			testName: "exampleURLWithTagsAndFilters",
			link: domain.Link{URL: "https://example.com/example",
				Tags: []string{"tag1", "tag2"}, Filters: []string{"filter1", "filter2"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			linkDTO := dto.LinkToLinkRequestDTO(&tc.link)

			payload, err := json.Marshal(&linkDTO)
			if err != nil {
				require.NoError(t, err)
			}

			linkUpdater := &mocks.LinkUpdater{}
			linkUpdater.On("UpdateLink", ctx, tgID, &tc.link).Return(nil)

			putLinksHandler := links.PutLinksHandler{LinkUpdater: linkUpdater}

			r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/links", bytes.NewReader(payload))
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

			w := httptest.NewRecorder()

			putLinksHandler.ServeHTTP(w, r)

			linkUpdater.AssertExpectations(t)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func Test_PutLinksHandler_ServeHTTP_InvalidTgID(t *testing.T) {
	ctx := context.Background()
	tgID := "invalidTgID"

	linkUpdater := &mocks.LinkUpdater{}
	putLinksHandler := links.PutLinksHandler{LinkUpdater: linkUpdater}

	r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/links", http.NoBody)
	r.Header.Set("Tg-Chat-Id", tgID)

	w := httptest.NewRecorder()

	putLinksHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse

	err := json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	if err != nil {
		require.NoError(t, err)
	}

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Invalid or missing tgID", *responseErrorBody.Description)
	assert.Equal(t, "400", *responseErrorBody.Code)
	assert.Equal(t, "INVALID_TG_ID", *responseErrorBody.ExceptionName)
}

func Test_PutLinksHandler_ServeHTTP_InvalidBody(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)

	linkUpdater := &mocks.LinkUpdater{}
	putLinksHandler := links.PutLinksHandler{LinkUpdater: linkUpdater}

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
			r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/links", tc.payload)
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

			w := httptest.NewRecorder()
			putLinksHandler.ServeHTTP(w, r)

			var responseErrorBody scrapperdto.ApiErrorResponse

			err := json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
			if err != nil {
				require.NoError(t, err)
			}

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Equal(t, "Invalid or missing request body", *responseErrorBody.Description)
			assert.Equal(t, "400", *responseErrorBody.Code)
			assert.Equal(t, "INVALID_REQUEST_BODY", *responseErrorBody.ExceptionName)
		})
	}
}

func Test_PutLinksHandler_ServeHTTP_InvalidBody_MissingRequiredFields(t *testing.T) {
	ctx := context.Background()
	linkUpdater := &mocks.LinkUpdater{}
	putLinksHandler := links.PutLinksHandler{LinkUpdater: linkUpdater}
	tgID := int64(123)

	payload := bytes.NewReader([]byte("{}"))

	r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/links", payload)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

	w := httptest.NewRecorder()

	putLinksHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse

	err := json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Missing required fields", *responseErrorBody.Description)
	assert.Equal(t, "400", *responseErrorBody.Code)
	assert.Equal(t, "INVALID_REQUEST_BODY", *responseErrorBody.ExceptionName)
}

func Test_PutLinksHandler_ServeHTTP_UpdateLinkError(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	link := domain.Link{URL: "https://example.com/example", Filters: []string{}, Tags: []string{}}
	linkDTO := dto.LinkToLinkRequestDTO(&link)

	payload, err := json.Marshal(&linkDTO)
	require.NoError(t, err)

	linkUpdater := &mocks.LinkUpdater{}
	linkUpdater.On("UpdateLink", ctx, tgID, &link).Return(errors.New("some error"))
	putLinksHandler := links.PutLinksHandler{LinkUpdater: linkUpdater}

	r := httptest.NewRequestWithContext(ctx, http.MethodPut, "/links", bytes.NewReader(payload))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

	w := httptest.NewRecorder()

	putLinksHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse

	err = json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Failed to update link", *responseErrorBody.Description)
	assert.Equal(t, "500", *responseErrorBody.Code)
	assert.Equal(t, "UPDATE_LINK_FAILED", *responseErrorBody.ExceptionName)
}
