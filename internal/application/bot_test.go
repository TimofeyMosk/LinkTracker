package application_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"LinkTracker/internal/application"
	"LinkTracker/internal/application/mocks"
	"LinkTracker/internal/domain"
)

const (
	commandTrack       = "/track"
	gitExampleURL      = "https://github.com/example/example"
	trackGoodResponse1 = "Введите адрес ссылки (поддерживается только gitHub и stackOverFlow)"
	trackGoodResponse2 = "Отправьте теги разделённые пробелами. Если не хотите добавлять теги введите \"-\" без кавычек"
	trackGoodResponse3 = "Отправьте фильтры разделённые пробелами. Если не хотите добавлять фильтры введите '-' без кавычек"
	trackGoodResponse4 = "Ссылка отслеживается"
)

func Test_Bot_HandleMessage_Start_Success(t *testing.T) {
	scrapper := &mocks.ScrapperClient{}
	bot := application.NewBot(scrapper)
	tgID := int64(123)
	text := "/start"

	scrapper.On("RegisterUser", tgID).Return(nil).Once()

	expectedText := "Добро пожаловать в LinkTracker, " +
		"это приложение для отслеживание изменений на github и stackoverflow." +
		"Для получения списка команд введите /help"

	responseText := bot.HandleMessage(tgID, text)

	assert.Equal(t, expectedText, responseText)
}

func Test_Bot_HandleMessage_Start_Error(t *testing.T) {
	scrapper := &mocks.ScrapperClient{}
	bot := application.NewBot(scrapper)
	tgID := int64(123)
	text := "/start"

	scrapper.On("RegisterUser", tgID).Return(errors.New("some error")).Once()

	expectedText := "Не удалось выполнить операцию"

	responseText := bot.HandleMessage(tgID, text)

	assert.Equal(t, expectedText, responseText)
}

func Test_Bot_HandleMessage_Help(t *testing.T) {
	scrapper := &mocks.ScrapperClient{}
	bot := application.NewBot(scrapper)
	tgID := int64(123)
	text := "/help"
	expectedText := "📝Доступные команды:\n\n" +
		"/start - Начало работы с ботом\n" +
		"/help - Помощь по командам\n" +
		"/track - Начать отслеживание ссылки\n" +
		"/untrack - Прекратить отслеживание\n" +
		"/list - Список отслеживаемых ссылок"

	responseText := bot.HandleMessage(tgID, text)

	assert.Equal(t, expectedText, responseText)
}

func Test_Bot_HandleMessage_Track(t *testing.T) {
	scrapper := &mocks.ScrapperClient{}
	bot := application.NewBot(scrapper)
	tgID := int64(123)
	message1 := commandTrack
	message2 := gitExampleURL
	message3 := "-"
	message4 := "-"
	expectedResponse1 := trackGoodResponse1
	expectedResponse2 := trackGoodResponse2
	expectedResponse3 := trackGoodResponse3
	expectedResponse4 := trackGoodResponse4

	scrapper.On("AddLink", tgID, domain.Link{URL: message2, Tags: []string{}, Filters: []string{}, ID: 0}).Return(nil).Once()

	response1 := bot.HandleMessage(tgID, message1)
	response2 := bot.HandleMessage(tgID, message2)
	response3 := bot.HandleMessage(tgID, message3)
	response4 := bot.HandleMessage(tgID, message4)

	assert.Equal(t, expectedResponse1, response1)
	assert.Equal(t, expectedResponse2, response2)
	assert.Equal(t, expectedResponse3, response3)
	assert.Equal(t, expectedResponse4, response4)
}

func Test_Bot_HandleMessage_Track_AlreadyExist(t *testing.T) {
	scrapper := &mocks.ScrapperClient{}
	bot := application.NewBot(scrapper)
	tgID := int64(123)
	message1 := commandTrack
	message2 := gitExampleURL
	message3 := "-"
	message4 := "-"
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
	errExpected := domain.ErrAPI{ExceptionMessage: domain.ErrLinkAlreadyTracking{}.Error()}

	scrapper.On("AddLink", tgID, domain.Link{URL: message2, Tags: []string{}, Filters: []string{}, ID: 0}).Return(nil).Once()
	scrapper.On("AddLink", tgID, domain.Link{URL: message2, Tags: []string{}, Filters: []string{}, ID: 0}).Return(errExpected).Once()

	response1 := bot.HandleMessage(tgID, message1)
	response2 := bot.HandleMessage(tgID, message2)
	response3 := bot.HandleMessage(tgID, message3)
	response4 := bot.HandleMessage(tgID, message4)
	response5 := bot.HandleMessage(tgID, message5)
	response6 := bot.HandleMessage(tgID, message6)
	response7 := bot.HandleMessage(tgID, message7)
	response8 := bot.HandleMessage(tgID, message8)

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
	scrapper := &mocks.ScrapperClient{}
	bot := application.NewBot(scrapper)
	tgID := int64(123)
	message1 := commandTrack
	message2 := "https://example.com/example/example"
	expectedResponse1 := trackGoodResponse1
	expectedResponse2 := "Поддерживается только gitHub(https://github.com/{owner}/{repo}) и " +
		"stackOverflow(https://stackoverflow.com/questions/{id}). Повторите команду /track"

	response1 := bot.HandleMessage(tgID, message1)
	response2 := bot.HandleMessage(tgID, message2)

	assert.Equal(t, expectedResponse1, response1)
	assert.Equal(t, expectedResponse2, response2)
}

func Test_Bot_HandleMessage_UnTrack(t *testing.T) {
	scrapper := &mocks.ScrapperClient{}
	bot := application.NewBot(scrapper)
	tgID := int64(123)
	message1 := "/untrack"
	message2 := gitExampleURL
	expectedResponse1 := "Введите адрес ссылки для удаления"
	expectedResponse2 := "Ссылка успешно удалена"

	scrapper.On("RemoveLink", tgID, domain.Link{URL: message2}).Return(nil).Once()

	response1 := bot.HandleMessage(tgID, message1)
	response2 := bot.HandleMessage(tgID, message2)

	assert.Equal(t, expectedResponse1, response1)
	assert.Equal(t, expectedResponse2, response2)
}

func Test_Bot_HandleMessage_UnTrack_Error(t *testing.T) {
	scrapper := &mocks.ScrapperClient{}
	bot := application.NewBot(scrapper)
	tgID := int64(123)
	message1 := "/untrack"
	message2 := gitExampleURL
	expectedResponse1 := "Введите адрес ссылки для удаления"
	expectedResponse2 := "Не удалось выполнить операцию"

	scrapper.On("RemoveLink", tgID, domain.Link{URL: message2}).Return(errors.New("some_errors")).Once()

	response1 := bot.HandleMessage(tgID, message1)
	response2 := bot.HandleMessage(tgID, message2)

	assert.Equal(t, expectedResponse1, response1)
	assert.Equal(t, expectedResponse2, response2)
}

func Test_Bot_HandleMessage_List(t *testing.T) {
	scrapper := &mocks.ScrapperClient{}
	bot := application.NewBot(scrapper)
	tgID := int64(123)
	message1 := "/list"

	scrapper.On("GetLinks", tgID).Return([]domain.Link{
		{URL: gitExampleURL, Tags: []string{}, Filters: []string{}, ID: 0},
		{URL: "https://github.com/example/example2", Tags: []string{"My", "Work"}, Filters: []string{"Me"}, ID: 1}}, nil).Once()

	response := bot.HandleMessage(tgID, message1)

	expectedResponse := "Список отслеживаемых ссылок:\n" +
		"https://github.com/example/example\n" +
		"https://github.com/example/example2 Tags: My Work  Filters: Me \n"

	assert.Equal(t, expectedResponse, response)
}
