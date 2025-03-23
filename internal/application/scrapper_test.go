package application_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"LinkTracker/internal/application"
	"LinkTracker/internal/application/mocks"
	"LinkTracker/internal/domain"
)

func Test_Scrapper_AddLink_Success(t *testing.T) {
	db := &mocks.Database{}
	bot := &mocks.BotClient{}
	gitClient := &mocks.GitHubClient{}
	soClient := &mocks.StackOverflowClient{}
	scrapper := application.NewScrapper(db, time.Minute, bot, gitClient, soClient)
	tgID := int64(123)
	newLink := domain.Link{URL: "https://github.com/example/example"}

	db.On("GetUserLinks", tgID).Return([]domain.Link{{URL: "https://github.com/example2/example2"}}, nil)
	db.On("AddLink", tgID, newLink).Return(nil)
	link, err := scrapper.AddLink(tgID, newLink)

	assert.NoError(t, err)
	assert.Equal(t, newLink, link)
}

func Test_Scrapper_AddLink_AlreadyTracking(t *testing.T) {
	db := &mocks.Database{}
	bot := &mocks.BotClient{}
	gitClient := &mocks.GitHubClient{}
	soClient := &mocks.StackOverflowClient{}
	scrapper := application.NewScrapper(db, time.Minute, bot, gitClient, soClient)
	tgID := int64(123)
	newLink := domain.Link{URL: "https://github.com/example/example"}

	db.On("GetUserLinks", tgID).Return([]domain.Link{{URL: "https://github.com/example2/example2"}, newLink}, nil)

	_, err := scrapper.AddLink(tgID, newLink)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrLinkAlreadyTracking{}, err)
}

func Test_Scrapper_AddLink_GetUserLinksError(t *testing.T) {
	db := &mocks.Database{}
	bot := &mocks.BotClient{}
	gitClient := &mocks.GitHubClient{}
	soClient := &mocks.StackOverflowClient{}
	scrapper := application.NewScrapper(db, time.Minute, bot, gitClient, soClient)
	tgID := int64(123)
	newLink := domain.Link{URL: "https://github.com/example/example"}

	db.On("GetUserLinks", tgID).Return(nil, errors.New("some error"))

	_, err := scrapper.AddLink(tgID, newLink)

	assert.Error(t, err)
}

func Test_Scrapper_AddLink_AddLinkError(t *testing.T) {
	db := &mocks.Database{}
	bot := &mocks.BotClient{}
	gitClient := &mocks.GitHubClient{}
	soClient := &mocks.StackOverflowClient{}
	scrapper := application.NewScrapper(db, time.Minute, bot, gitClient, soClient)
	tgID := int64(123)
	newLink := domain.Link{URL: "https://github.com/example/example"}

	db.On("GetUserLinks", tgID).Return([]domain.Link{{URL: "https://github.com/example2/example2"}}, nil)
	db.On("AddLink", tgID, newLink).Return(errors.New("some error"))

	_, err := scrapper.AddLink(tgID, newLink)

	assert.Error(t, err)
}

func Test_Scrapper_Scrape(t *testing.T) {
	db := &mocks.Database{}
	bot := &mocks.BotClient{}
	gitClient := &mocks.GitHubClient{}
	soClient := &mocks.StackOverflowClient{}
	scrapper := application.NewScrapper(db, time.Minute, bot, gitClient, soClient)
	tgID1 := int64(1)
	tgID2 := int64(2)
	link1 := domain.Link{URL: "https://github.com/example1/example1"}
	link2 := domain.Link{URL: "https://stackoverflow.com/example2/example2"}

	db.On("GetAllUsers").Return([]int64{tgID1, tgID2}, nil).Once()
	db.On("GetUserLinks", tgID1).Return([]domain.Link{link1}, nil).Once()
	db.On("GetUserLinks", tgID2).Return([]domain.Link{link2}, nil).Once()
	gitClient.On("GetLastUpdateTimeRepo", link1.URL).Return(time.Now(), nil).Once()
	soClient.On("GetLastActivityQuestion", link2.URL).Return(time.Now(), nil).Once()
	bot.On("PostUpdates", link1, tgID1).Return(nil).Once()
	bot.On("PostUpdates", link2, tgID2).Return(nil).Once()

	scrapper.Scrape()

	bot.AssertExpectations(t)
	gitClient.AssertExpectations(t)
	soClient.AssertExpectations(t)
	db.AssertExpectations(t)
}
