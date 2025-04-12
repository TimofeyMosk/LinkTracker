package linkchecker_test

// Test_LinkChecker_CheckLinks проверяет корректность работы метода CheckLinks.
// func Test_LinkChecker_CheckLinks(t *testing.T) {
//	ctx := context.Background()
//	linkRepoMock := &scrappermocks.LinkRepo{}
//	sourceHandlerMock := &mocks.LinkSourceHandler{}
//	limitLinksInPage := int64(10)
//	workers := 4
//	links := []domain.Link{{ID: 1,
//		URL:  "https://github.com/central-university-dev/central-university-dev",
//		Tags: []string{}, Filters: []string{},
//		LastUpdated: time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC)}}
//
//	linkChecker := linkchecker.NewLinkChecker(linkRepoMock, []linkchecker.LinkSourceHandler{sourceHandlerMock},
//		limitLinksInPage, workers)
//	linkUpdates := make(chan domain.LinkUpdate)
//
//	sourceHandlerMock.On("Check", mock.Anything).Return(true, nil)
//
//	linkRepoMock.On("GetLinksAfter", ctx, time.Time{}, limitLinksInPage).Return(links, nil).Once()
//	linkRepoMock.On("GetLinksAfter", ctx, time.Time{}, limitLinksInPage).Return([]domain.Link{}, nil).Once()
//
//	linkChecker.CheckLinks(ctx, linkUpdates)
//}
