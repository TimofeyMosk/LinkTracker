package linkchecker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"LinkTracker/internal/application/scrapper/linkchecker"
	"LinkTracker/internal/application/scrapper/linkchecker/mocks"
	scrappermocks "LinkTracker/internal/application/scrapper/mocks"
	"LinkTracker/internal/domain"
)

// Test_LinkChecker_CheckLinks проверяет корректность работы метода CheckLinks.
func Test_LinkChecker_CheckLinks(t *testing.T) {
	ctx := context.Background()
	linkRepo := &scrappermocks.LinkRepo{}
	handler := &mocks.LinkSourceHandler{}
	limitLinksInPage := int64(500)
	workers := 4
	linkUpdates := make(chan domain.LinkUpdate, 100)
	updateTime := time.Now()
	usersTgIDs := []int64{1, 2, 3}
	descriptionUpdate := "update"

	link1 := domain.Link{URL: "https://example/example", ID: 1,
		LastUpdated: time.Date(2025, 1, 1, 1, 1, 1, 1, time.UTC)}
	link2 := domain.Link{URL: "https://example/example2", ID: 2,
		LastUpdated: time.Date(2025, 2, 2, 2, 2, 2, 2, time.UTC)}
	links := []domain.Link{link1, link2}

	linkRepo.On("GetLinksAfter", ctx, time.Time{}, limitLinksInPage).Return(links, nil).Once()
	linkRepo.On("GetLinksAfter", ctx, link2.LastUpdated, limitLinksInPage).Return(nil, nil).Once()
	handler.On("Supports", mock.Anything).Return(true)
	handler.On("Check", ctx, &link1).Return(time.Time{}, "", errors.New("not Updates")).Once()
	handler.On("Check", ctx, &link2).Return(updateTime, descriptionUpdate, nil).Once()
	linkRepo.On("UpdateTimeLink", ctx, updateTime, link2.ID).Return(nil)
	linkRepo.On("GetUsersByLink", ctx, link2.ID).Return(usersTgIDs, nil)

	linksChecker := linkchecker.NewLinkChecker(linkRepo, []linkchecker.LinkSourceHandler{handler}, limitLinksInPage, workers)

	linksChecker.CheckLinks(ctx, linkUpdates)

	update2 := <-linkUpdates

	assert.Equal(t, usersTgIDs, update2.TgIDs)
	assert.Equal(t, link2, update2.Link)
	assert.Equal(t, descriptionUpdate, update2.Description)

	linkRepo.AssertExpectations(t)
	handler.AssertExpectations(t)
}
