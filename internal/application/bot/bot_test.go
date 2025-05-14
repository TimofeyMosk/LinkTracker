package bot_test

import (
	"context"
	"errors"
	"testing"

	"LinkTracker/internal/application/bot"
	"LinkTracker/internal/domain"

	"github.com/stretchr/testify/assert"

	"LinkTracker/internal/application/bot/mocks"
)

const (
	WaitingLink = iota
	WaitingTags
	WaitingFilters
	WaitingDelete
	WaitingSetTagsWaitingLink
	WaitingSetTagsWaitingTags
	commandTrack       = "/track"
	gitExampleURL      = "https://github.com/example/example"
	trackGoodResponse1 = "Введите адрес ссылки (поддерживается только gitHub и stackOverFlow)"
	trackGoodResponse2 = "Отправьте теги разделённые пробелами. Если не хотите добавлять теги введите \"-\" без кавычек"
	trackGoodResponse3 = "Отправьте фильтры разделённые пробелами. Если не хотите добавлять фильтры введите '-' без кавычек"
	trackGoodResponse4 = "Ссылка отслеживается"
	oneTag             = "tag"
)

func Test_Bot_HandleMessage_Start_Success(t *testing.T) {
	ctx := context.Background()
	scrapper := &mocks.ScrapperClient{}
	tgClient := &mocks.TelegramClient{}
	Bot := bot.NewBot(scrapper, tgClient)
	tgID := int64(123)
	text := "/start"

	scrapper.On("RegisterUser", ctx, tgID).Return(nil).Once()

	expectedText := "Добро пожаловать в LinkTracker, " +
		"это приложение для отслеживание изменений на github и stackoverflow." +
		"Для получения списка команд введите /help"

	responseText := Bot.HandleMessage(ctx, tgID, text)

	assert.Equal(t, expectedText, responseText)
}

func Test_Bot_HandleMessage_Start_Error(t *testing.T) {
	ctx := context.Background()
	scrapper := &mocks.ScrapperClient{}
	tgClient := &mocks.TelegramClient{}
	Bot := bot.NewBot(scrapper, tgClient)
	tgID := int64(123)
	text := "/start"

	scrapper.On("RegisterUser", ctx, tgID).Return(errors.New("some error")).Once()

	expectedText := "Не удалось выполнить операцию. Возможно, вы уже зарегистрированы в приложении"

	responseText := Bot.HandleMessage(ctx, tgID, text)

	assert.Equal(t, expectedText, responseText)
}

func Test_Bot_HandleMessage_Help(t *testing.T) {
	ctx := context.Background()
	scrapper := &mocks.ScrapperClient{}
	tgClient := &mocks.TelegramClient{}
	Bot := bot.NewBot(scrapper, tgClient)
	tgID := int64(123)
	text := "/help"
	expectedText := "📝Доступные команды:\n\n" +
		"/start - Начало работы с ботом\n" +
		"/help - Помощь по командам\n" +
		"/track - Начать отслеживание ссылки\n" +
		"/untrack - Прекратить отслеживание\n" +
		"/list - Список отслеживаемых ссылок"

	responseText := Bot.HandleMessage(ctx, tgID, text)

	assert.Equal(t, expectedText, responseText)
}

func Test_Bot_HandleMessage_Track(t *testing.T) {
	ctx := context.Background()
	scrapper := &mocks.ScrapperClient{}
	tgClient := &mocks.TelegramClient{}
	Bot := bot.NewBot(scrapper, tgClient)

	tgID := int64(123)
	message1 := commandTrack
	message2 := gitExampleURL
	message3 := oneTag
	message4 := "filter"
	expectedResponse1 := trackGoodResponse1
	expectedResponse2 := trackGoodResponse2
	expectedResponse3 := trackGoodResponse3
	expectedResponse4 := trackGoodResponse4
	emptyLink := domain.Link{}
	linkWithURL := domain.Link{URL: gitExampleURL}
	linkWithTags := domain.Link{URL: gitExampleURL, Tags: []string{oneTag}}
	linkWithFilters := domain.Link{URL: gitExampleURL, Tags: []string{oneTag}, Filters: []string{"filter"}}

	scrapper.On("CreateState", ctx, tgID, WaitingLink).Return(nil).Once()
	scrapper.On("GetState", ctx, tgID).Return(WaitingLink, emptyLink, nil).Once()
	scrapper.On("UpdateState", ctx, tgID, WaitingTags, &linkWithURL).Return(nil).Once()
	scrapper.On("GetState", ctx, tgID).Return(WaitingTags, linkWithURL, nil).Once()
	scrapper.On("UpdateState", ctx, tgID, WaitingFilters, &linkWithTags).Return(nil).Once()
	scrapper.On("GetState", ctx, tgID).Return(WaitingFilters, linkWithTags, nil).Once()
	scrapper.On("DeleteState", ctx, tgID).Return(nil).Once()
	scrapper.On("AddLink", ctx, tgID, &linkWithFilters).Return(nil).Once()

	response1 := Bot.HandleMessage(ctx, tgID, message1)
	response2 := Bot.HandleMessage(ctx, tgID, message2)
	response3 := Bot.HandleMessage(ctx, tgID, message3)
	response4 := Bot.HandleMessage(ctx, tgID, message4)

	assert.Equal(t, expectedResponse1, response1)
	assert.Equal(t, expectedResponse2, response2)
	assert.Equal(t, expectedResponse3, response3)
	assert.Equal(t, expectedResponse4, response4)
}

func Test_Bot_HandleMessage_Track_AlreadyExist(t *testing.T) {
	ctx := context.Background()
	scrapper := &mocks.ScrapperClient{}
	tgClient := &mocks.TelegramClient{}
	Bot := bot.NewBot(scrapper, tgClient)

	tgID := int64(123)
	message1 := commandTrack
	message2 := gitExampleURL
	message3 := oneTag
	message4 := "filter"
	message5 := commandTrack
	message6 := gitExampleURL
	message7 := "-"
	message8 := "-"
	expectedResponse1 := trackGoodResponse1
	expectedResponse2 := trackGoodResponse2
	expectedResponse3 := trackGoodResponse3
	expectedResponse4 := trackGoodResponse4
	expectedResponse5 := trackGoodResponse1
	expectedResponse6 := trackGoodResponse2
	expectedResponse7 := trackGoodResponse3
	expectedResponse8 := "Данная ссылка уже отслеживается"
	emptyLink := domain.Link{}
	linkWithURL := domain.Link{URL: gitExampleURL}
	linkWithTags := domain.Link{URL: gitExampleURL, Tags: []string{oneTag}}
	linkWithoutTags := domain.Link{URL: gitExampleURL, Tags: []string{}}
	linkWithFilters := domain.Link{URL: gitExampleURL, Tags: []string{oneTag}, Filters: []string{"filter"}}
	linkWithoutFilters := domain.Link{URL: gitExampleURL, Tags: []string{}, Filters: []string{}}
	errExpected := domain.ErrAPI{ExceptionMessage: domain.ErrLinkAlreadyTracking{}.Error()}

	scrapper.On("CreateState", ctx, tgID, WaitingLink).Return(nil).Once()
	scrapper.On("GetState", ctx, tgID).Return(WaitingLink, emptyLink, nil).Once()
	scrapper.On("UpdateState", ctx, tgID, WaitingTags, &linkWithURL).Return(nil).Once()
	scrapper.On("GetState", ctx, tgID).Return(WaitingTags, linkWithURL, nil).Once()
	scrapper.On("UpdateState", ctx, tgID, WaitingFilters, &linkWithTags).Return(nil).Once()
	scrapper.On("GetState", ctx, tgID).Return(WaitingFilters, linkWithTags, nil).Once()
	scrapper.On("DeleteState", ctx, tgID).Return(nil).Once()
	scrapper.On("AddLink", ctx, tgID, &linkWithFilters).Return(nil).Once()

	scrapper.On("CreateState", ctx, tgID, WaitingLink).Return(nil).Once()
	scrapper.On("GetState", ctx, tgID).Return(WaitingLink, emptyLink, nil).Once()
	scrapper.On("UpdateState", ctx, tgID, WaitingTags, &linkWithURL).Return(nil).Once()
	scrapper.On("GetState", ctx, tgID).Return(WaitingTags, linkWithURL, nil).Once()
	scrapper.On("UpdateState", ctx, tgID, WaitingFilters, &linkWithoutTags).Return(nil).Once()
	scrapper.On("GetState", ctx, tgID).Return(WaitingFilters, linkWithoutTags, nil).Once()
	scrapper.On("DeleteState", ctx, tgID).Return(nil).Once()
	scrapper.On("AddLink", ctx, tgID, &linkWithoutFilters).Return(errExpected).Once()

	response1 := Bot.HandleMessage(ctx, tgID, message1)
	response2 := Bot.HandleMessage(ctx, tgID, message2)
	response3 := Bot.HandleMessage(ctx, tgID, message3)
	response4 := Bot.HandleMessage(ctx, tgID, message4)
	response5 := Bot.HandleMessage(ctx, tgID, message5)
	response6 := Bot.HandleMessage(ctx, tgID, message6)
	response7 := Bot.HandleMessage(ctx, tgID, message7)
	response8 := Bot.HandleMessage(ctx, tgID, message8)

	assert.Equal(t, expectedResponse1, response1)
	assert.Equal(t, expectedResponse2, response2)
	assert.Equal(t, expectedResponse3, response3)
	assert.Equal(t, expectedResponse4, response4)
	assert.Equal(t, expectedResponse5, response5)
	assert.Equal(t, expectedResponse6, response6)
	assert.Equal(t, expectedResponse7, response7)
	assert.Equal(t, expectedResponse8, response8)
}

func Test_Bot_HandleMessage_Track_InvalidLink(t *testing.T) {
	ctx := context.Background()
	scrapper := &mocks.ScrapperClient{}
	tgClient := &mocks.TelegramClient{}
	Bot := bot.NewBot(scrapper, tgClient)

	tgID := int64(123)
	message1 := commandTrack
	message2 := "https://example.com/example/example"
	expectedResponse1 := trackGoodResponse1
	expectedResponse2 := "Поддерживается только gitHub(https://github.com/{owner}/{repo}) и " +
		"stackOverflow(https://stackoverflow.com/questions/{id}). Повторите команду /track"
	emptyLink := domain.Link{}

	scrapper.On("CreateState", ctx, tgID, WaitingLink).Return(nil).Once()
	scrapper.On("GetState", ctx, tgID).Return(WaitingLink, emptyLink, nil).Once()
	scrapper.On("DeleteState", ctx, tgID).Return(nil).Once()
	response1 := Bot.HandleMessage(ctx, tgID, message1)
	response2 := Bot.HandleMessage(ctx, tgID, message2)

	assert.Equal(t, expectedResponse1, response1)
	assert.Equal(t, expectedResponse2, response2)
}

func Test_Bot_HandleMessage_UnTrack(t *testing.T) {
	ctx := context.Background()
	scrapper := &mocks.ScrapperClient{}
	tgClient := &mocks.TelegramClient{}
	Bot := bot.NewBot(scrapper, tgClient)

	tgID := int64(123)
	message1 := "/untrack"
	message2 := gitExampleURL
	expectedResponse1 := "Введите адрес ссылки для удаления"
	expectedResponse2 := "Ссылка успешно удалена"
	emptyLink := domain.Link{}
	linkWithURL := domain.Link{URL: gitExampleURL}

	scrapper.On("CreateState", ctx, tgID, WaitingDelete).Return(nil).Once()
	scrapper.On("GetState", ctx, tgID).Return(WaitingDelete, emptyLink, nil).Once()
	scrapper.On("DeleteState", ctx, tgID).Return(nil).Once()
	scrapper.On("RemoveLink", ctx, tgID, &linkWithURL).Return(nil).Once()

	response1 := Bot.HandleMessage(ctx, tgID, message1)
	response2 := Bot.HandleMessage(ctx, tgID, message2)

	assert.Equal(t, expectedResponse1, response1)
	assert.Equal(t, expectedResponse2, response2)
}

func Test_Bot_HandleMessage_UnTrack_Error(t *testing.T) {
	ctx := context.Background()
	scrapper := &mocks.ScrapperClient{}
	tgClient := &mocks.TelegramClient{}
	Bot := bot.NewBot(scrapper, tgClient)

	tgID := int64(123)
	message1 := "/untrack"
	message2 := gitExampleURL
	expectedResponse1 := "Введите адрес ссылки для удаления"
	expectedResponse2 := "Не удалось выполнить операцию"
	emptyLink := domain.Link{}
	linkWithURL := domain.Link{URL: gitExampleURL}

	scrapper.On("CreateState", ctx, tgID, WaitingDelete).Return(nil).Once()
	scrapper.On("GetState", ctx, tgID).Return(WaitingDelete, emptyLink, nil).Once()
	scrapper.On("DeleteState", ctx, tgID).Return(nil).Once()
	scrapper.On("RemoveLink", ctx, tgID, &linkWithURL).Return(errors.New("some_errors")).Once()

	response1 := Bot.HandleMessage(ctx, tgID, message1)
	response2 := Bot.HandleMessage(ctx, tgID, message2)

	assert.Equal(t, expectedResponse1, response1)
	assert.Equal(t, expectedResponse2, response2)
}

func Test_Bot_HandleMessage_List(t *testing.T) {
	ctx := context.Background()
	scrapper := &mocks.ScrapperClient{}
	tgClient := &mocks.TelegramClient{}
	Bot := bot.NewBot(scrapper, tgClient)

	tgID := int64(123)
	message1 := "/list"

	scrapper.On("GetLinks", ctx, tgID).Return([]domain.Link{
		{URL: gitExampleURL, Tags: []string{}, Filters: []string{}, ID: 0},
		{URL: "https://github.com/example/example2", Tags: []string{"Work"}, Filters: []string{"My"}, ID: 1}}, nil).Once()

	response := Bot.HandleMessage(ctx, tgID, message1)

	expectedResponse := "Список отслеживаемых ссылок:\n" +
		"Work: \n" +
		"linkID: 1 Url: https://github.com/example/example2 Tags: Work Filters: My\n" +
		"\nБез тегов: \n" +
		"linkID: 0 " + "Url: " + gitExampleURL + "\n"

	assert.Equal(t, expectedResponse, response)
}

func Test_Bot_HandleMessage_SetTags(t *testing.T) {
	ctx := context.Background()
	scrapper := &mocks.ScrapperClient{}
	tgClient := &mocks.TelegramClient{}
	Bot := bot.NewBot(scrapper, tgClient)

	tgID := int64(123)
	message1 := "/settags"
	message2 := gitExampleURL
	message3 := oneTag
	emptyLink := domain.Link{}
	linkWithURL := domain.Link{URL: gitExampleURL, Tags: []string{}, Filters: []string{}, ID: 0}
	linkWithTags := domain.Link{URL: gitExampleURL, Tags: []string{oneTag}, Filters: []string{}, ID: 0}

	scrapper.On("CreateState", ctx, tgID, WaitingSetTagsWaitingLink).Return(nil).Once()
	scrapper.On("GetState", ctx, tgID).Return(WaitingSetTagsWaitingLink, emptyLink, nil).Once()
	scrapper.On("GetLinks", ctx, tgID).Return([]domain.Link{linkWithURL}, nil)
	scrapper.On("UpdateState", ctx, tgID, WaitingSetTagsWaitingTags, &linkWithURL).Return(nil).Once()
	scrapper.On("GetState", ctx, tgID).Return(WaitingSetTagsWaitingTags, linkWithURL, nil).Once()
	scrapper.On("DeleteState", ctx, tgID).Return(nil).Once()
	scrapper.On("UpdateLink", ctx, tgID, &linkWithTags).Return(nil).Once()

	response1 := Bot.HandleMessage(ctx, tgID, message1)
	response2 := Bot.HandleMessage(ctx, tgID, message2)
	response3 := Bot.HandleMessage(ctx, tgID, message3)

	expectedResponse1 := "Введите ссылку, для которой хотите изменить тег/теги"
	expectedResponse2 := "Отправьте новые теги разделённые пробелами. Если не хотите добавлять теги отправьте '-' без кавычек"
	expectedResponse3 := "Теги успешно изменены"

	assert.Equal(t, expectedResponse1, response1)
	assert.Equal(t, expectedResponse2, response2)
	assert.Equal(t, expectedResponse3, response3)
}
