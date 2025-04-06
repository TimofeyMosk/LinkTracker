package scrapper_test

// import (
//	"LinkTracker/internal/application"
//	"LinkTracker/internal/application/mocks"
//	"LinkTracker/internal/domain"
//	"context"
//	"errors"
//	"github.com/stretchr/testify/assert"
//	"testing"
//	"time"
//)
//
// func Test_Scrapper_AddLink_Success(t *testing.T) {
//	linkRepo := &mocks.LinkRepo{}
//	userRepo := &mocks.UserRepo{}
//	bot := &mocks.BotClient{}
//	gitClient := &mocks.GitHubClient{}
//	soClient := &mocks.StackOverflowClient{}
//	ctx := context.Background()
//	scrapper := application.NewScrapper(ctx, userRepo, linkRepo, time.Minute, bot, gitClient, soClient)
//	tgID := int64(123)
//	newLink := domain.Link{URL: "https://github.com/example/example"}
//
//	linkRepo.On("GetUserLinks", ctx, tgID).Return([]domain.Link{{URL: "https://github.com/example2/example2"}}, nil)
//	linkRepo.On("AddLink", ctx, tgID, newLink).Return(nil)
//	link, err := scrapper.AddLink(tgID, newLink)
//
//	assert.NoError(t, err)
//	assert.Equal(t, newLink, link)
//}
//
// func Test_Scrapper_AddLink_AlreadyTracking(t *testing.T) {
//	linkRepo := &mocks.LinkRepo{}
//	userRepo := &mocks.UserRepo{}
//	bot := &mocks.BotClient{}
//	gitClient := &mocks.GitHubClient{}
//	soClient := &mocks.StackOverflowClient{}
//
//	ctx := context.Background()
//	scrapper := application.NewScrapper(ctx, userRepo, linkRepo, time.Minute, bot, gitClient, soClient)
//	tgID := int64(123)
//	newLink := domain.Link{URL: "https://github.com/example/example"}
//
//	linkRepo.On("GetUserLinks", ctx, tgID).Return([]domain.Link{{URL: "https://github.com/example2/example2"}, newLink}, nil)
//
//	_, err := scrapper.AddLink(tgID, newLink)
//
//	assert.Error(t, err)
//	assert.Equal(t, domain.ErrLinkAlreadyTracking{}, err)
//}
//
// func Test_Scrapper_AddLink_GetUserLinksError(t *testing.T) {
//	linkRepo := &mocks.LinkRepo{}
//	userRepo := &mocks.UserRepo{}
//	bot := &mocks.BotClient{}
//	gitClient := &mocks.GitHubClient{}
//	soClient := &mocks.StackOverflowClient{}
//	ctx := context.Background()
//	scrapper := application.NewScrapper(ctx, userRepo, linkRepo, time.Minute, bot, gitClient, soClient)
//	tgID := int64(123)
//	newLink := domain.Link{URL: "https://github.com/example/example"}
//
//	linkRepo.On("GetUserLinks", ctx, tgID).Return(nil, errors.New("some error"))
//
//	_, err := scrapper.AddLink(tgID, newLink)
//
//	assert.Error(t, err)
//}
//
// func Test_Scrapper_AddLink_AddLinkError(t *testing.T) {
//	linkRepo := &mocks.LinkRepo{}
//	userRepo := &mocks.UserRepo{}
//	bot := &mocks.BotClient{}
//	gitClient := &mocks.GitHubClient{}
//	soClient := &mocks.StackOverflowClient{}
//	ctx := context.Background()
//	scrapper := application.NewScrapper(ctx, userRepo, linkRepo, time.Minute, bot, gitClient, soClient)
//	tgID := int64(123)
//	newLink := domain.Link{URL: "https://github.com/example/example"}
//
//	linkRepo.On("GetUserLinks", ctx, tgID).Return([]domain.Link{{URL: "https://github.com/example2/example2"}}, nil)
//	linkRepo.On("AddLink", ctx, tgID, newLink).Return(errors.New("some error"))
//
//	_, err := scrapper.AddLink(tgID, newLink)
//
//	assert.Error(t, err)
//}
//
// func Test_Scrapper_Scrape(t *testing.T) {
//	linkRepo := &mocks.LinkRepo{}
//	userRepo := &mocks.UserRepo{}
//	bot := &mocks.BotClient{}
//	gitClient := &mocks.GitHubClient{}
//	soClient := &mocks.StackOverflowClient{}
//	ctx := context.Background()
//	scrapper := application.NewScrapper(ctx, userRepo, linkRepo, time.Minute, bot, gitClient, soClient)
//	tgID1 := int64(3)
//	tgID2 := int64(4)
//	link1 := domain.Link{URL: "https://github.com/example1/example1", ID: 1, LastUpdated: time.Now().Add(-10 * time.Hour)}
//	link2 := domain.Link{URL: "https://stackoverflow.com/example2/example2", ID: 2, LastUpdated: time.Now().Add(-10 * time.Hour)}
//	newTimeLink1 := time.Now()
//	newTimeLink2 := time.Now()
//
//	linkRepo.On("GetAllLinks", ctx).Return([]domain.Link{link1, link2}, nil).Once()
//	linkRepo.On("GetUsersByLink", ctx, link1.ID).Return([]int64{tgID1}, nil).Once()
//	linkRepo.On("GetUsersByLink", ctx, link2.ID).Return([]int64{tgID1, tgID2}, nil).Once()
//	linkRepo.On("UpdateTimeLink", ctx, newTimeLink1, link1.ID).Return(nil)
//	linkRepo.On("UpdateTimeLink", ctx, newTimeLink2, link2.ID).Return(nil)
//	gitClient.On("GetLatestPROrIssue", link1.URL).Return(newTimeLink1, "description1", nil).Once()
//	soClient.On("GetLatestAnswerOrComment", link2.URL).Return(newTimeLink2, "description2", nil).Once()
//	bot.On("PostUpdates", link1, []int64{tgID1}, "description1").Return(nil).Once()
//	bot.On("PostUpdates", link2, []int64{tgID1, tgID2}, "description2").Return(nil).Once()
//
//	scrapper.Scrape()
//
//	bot.AssertExpectations(t)
//	gitClient.AssertExpectations(t)
//	soClient.AssertExpectations(t)
//	userRepo.AssertExpectations(t)
//}
