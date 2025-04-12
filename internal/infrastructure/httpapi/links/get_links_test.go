package links_test

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
	"LinkTracker/internal/infrastructure/httpapi/links"
	"LinkTracker/internal/infrastructure/httpapi/links/mocks"
)

func Test_GetLinksHandler_InvalidChatID(t *testing.T) {
	ctx := context.Background()
	tgID := "invalidTgID"
	linkGetter := &mocks.LinkGetter{}
	postLinksHandler := links.GetLinksHandler{LinkGetter: linkGetter}

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

func Test_GetLinksHandler_UserNotExist(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)

	linkGetter := &mocks.LinkGetter{}
	linkGetter.On("GetUserLinks", ctx, tgID).Return([]domain.Link{}, domain.ErrUserNotExist{})
	postLinksHandler := links.GetLinksHandler{LinkGetter: linkGetter}

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/links", http.NoBody)
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	postLinksHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "User not exist", *responseErrorBody.Description)
	assert.Equal(t, "404", *responseErrorBody.Code)
	assert.Equal(t, "USER_NOT_EXIST", *responseErrorBody.ExceptionName)
}

func Test_GetLinksHandler_LinksNotReceived(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)

	linkGetter := &mocks.LinkGetter{}
	linkGetter.On("GetUserLinks", ctx, tgID).Return([]domain.Link{}, errors.New("some error"))
	postLinksHandler := links.GetLinksHandler{LinkGetter: linkGetter}

	r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/links", http.NoBody)
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	postLinksHandler.ServeHTTP(w, r)

	var responseErrorBody scrapperdto.ApiErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &responseErrorBody)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Links not received", *responseErrorBody.Description)
	assert.Equal(t, "500", *responseErrorBody.Code)
	assert.Equal(t, "LINKS_NOT_RECEIVED", *responseErrorBody.ExceptionName)
}

func Test_GetLinksHandler_Success(t *testing.T) {
	ctx := context.Background()
	tgID := int64(123)
	link1 := domain.Link{URL: "https://example/example",
		Tags: []string{"tag1", "tag2"}, Filters: []string{"filter1", "filter2"}, ID: 1}
	link2 := domain.Link{URL: "https://example2/example2", ID: 2}
	linksSlice := []domain.Link{link1, link2}

	linkGetter := &mocks.LinkGetter{}
	linkGetter.On("GetUserLinks", ctx, tgID).Return(linksSlice, nil)
	postLinksHandler := links.GetLinksHandler{LinkGetter: linkGetter}

	r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/links", http.NoBody)
	r.Header.Set("Tg-Chat-Id", strconv.FormatInt(tgID, 10))

	w := httptest.NewRecorder()

	postLinksHandler.ServeHTTP(w, r)

	var listLinksResponse scrapperdto.ListLinksResponse
	err := json.Unmarshal(w.Body.Bytes(), &listLinksResponse)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, len(linksSlice), int(*listLinksResponse.Size))
	require.NotNil(t, listLinksResponse.Links)
	assert.Equal(t, link1.URL, *(*listLinksResponse.Links)[0].Url)
	assert.Equal(t, link1.ID, *(*listLinksResponse.Links)[0].Id)
	assert.Equal(t, link1.Tags, *(*listLinksResponse.Links)[0].Tags)
	assert.Equal(t, link1.Filters, *(*listLinksResponse.Links)[0].Filters)
	assert.Equal(t, link2.URL, *(*listLinksResponse.Links)[1].Url)
	assert.Equal(t, link2.ID, *(*listLinksResponse.Links)[1].Id)
}
