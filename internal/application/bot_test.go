package application_test

import (
	"errors"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/mock"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/mocks"
)

func TestBot_Start_CommandStart_Success(t *testing.T) {
	mockTelegram := &mocks.TelegramClient{}
	mockScrapper := &mocks.ScrapperClient{}

	bot := application.NewBot(mockScrapper, mockTelegram, 8)
	command := "/start"
	updates := make(chan tgbotapi.Update, 1)
	updates <- tgbotapi.Update{
		Message: &tgbotapi.Message{Text: command, Chat: &tgbotapi.Chat{ID: 123}, Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: len(command)}}},
	}
	close(updates)
	mockTelegram.On("GetUpdates").Return(tgbotapi.UpdatesChannel(updates)).Once()
	mockScrapper.On("RegisterUser", mock.AnythingOfType("int64")).Return(nil).Once()
	mockTelegram.On("SendMessage", mock.AnythingOfType("int64"), "Добро пожаловать в LinkTracker, "+
		"это приложение для отслеживание изменений на github и stackoverflow."+
		"Для получения списка команд введите /help").Return(tgbotapi.Message{}, nil).Once()

	bot.Start()
	mockTelegram.AssertExpectations(t)
	mockScrapper.AssertExpectations(t)
}

func TestBot_Start_CommandStart_RegisterError(t *testing.T) {
	mockTelegram := &mocks.TelegramClient{}
	mockScrapper := &mocks.ScrapperClient{}
	bot := application.NewBot(mockScrapper, mockTelegram, 8)
	command := "/start"
	updates := make(chan tgbotapi.Update, 1)
	updates <- tgbotapi.Update{
		Message: &tgbotapi.Message{Text: command, Chat: &tgbotapi.Chat{ID: 123}, Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: len(command)}}},
	}
	close(updates)
	mockTelegram.On("GetUpdates").Return(tgbotapi.UpdatesChannel(updates)).Once()
	mockScrapper.On("RegisterUser", mock.AnythingOfType("int64")).Return(errors.New("error")).Once()
	mockTelegram.On("SendMessage", mock.AnythingOfType("int64"), "Не удалось выполнить операцию").Return(tgbotapi.Message{}, nil).Once()

	bot.Start()
	mockTelegram.AssertExpectations(t)
	mockScrapper.AssertExpectations(t)
}

func TestBot_Start_CommandHelp(t *testing.T) {
	mockTelegram := &mocks.TelegramClient{}
	mockScrapper := &mocks.ScrapperClient{}
	bot := application.NewBot(mockScrapper, mockTelegram, 8)
	command := "/help"
	updates := make(chan tgbotapi.Update, 1)
	updates <- tgbotapi.Update{
		Message: &tgbotapi.Message{Text: command, Chat: &tgbotapi.Chat{ID: 123}, Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: len(command)}}},
	}
	close(updates)
	mockTelegram.On("GetUpdates").Return(tgbotapi.UpdatesChannel(updates)).Once()
	mockTelegram.On("SendMessage", mock.AnythingOfType("int64"), "📝Доступные команды:\n\n"+
		"/start - Начало работы с ботом\n"+
		"/help - Помощь по командам\n"+
		"/track - Начать отслеживание ссылки\n"+
		"/untrack - Прекратить отслеживание\n"+
		"/list - Список отслеживаемых ссылок\n").Return(tgbotapi.Message{}, nil).Once()

	bot.Start()

	mockTelegram.AssertExpectations(t)
	mockScrapper.AssertExpectations(t)
}

func TestBot_Start_CommandTrack(t *testing.T) {
	mockTelegram := &mocks.TelegramClient{}
	mockScrapper := &mocks.ScrapperClient{}
	bot := application.NewBot(mockScrapper, mockTelegram, 8)
	command := "/track"
	linkURL := "https://github.com/TimofeyMosk/fractalFlame-image-creator"
	updates := make(chan tgbotapi.Update, 100)
	updates <- tgbotapi.Update{
		Message: &tgbotapi.Message{Text: command, Chat: &tgbotapi.Chat{ID: 123}, Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: len(command)}}},
	}
	updates <- tgbotapi.Update{
		Message: &tgbotapi.Message{Text: linkURL, Chat: &tgbotapi.Chat{ID: 123}},
	}
	updates <- tgbotapi.Update{
		Message: &tgbotapi.Message{Text: "-", Chat: &tgbotapi.Chat{ID: 123}},
	}
	updates <- tgbotapi.Update{
		Message: &tgbotapi.Message{Text: "-", Chat: &tgbotapi.Chat{ID: 123}},
	}
	close(updates)

	mockTelegram.On("GetUpdates").Return(tgbotapi.UpdatesChannel(updates)).Once()
	mockTelegram.On("SendMessage", mock.AnythingOfType("int64"),
		"Введите адрес ссылки (поддерживается только gitHub и stackOverFlow").Return(tgbotapi.Message{}, nil).Once()
	mockTelegram.On("SendMessage", mock.AnythingOfType("int64"),
		"Отправьте теги разделённые пробелами. Если не хотите добавлять теги введите \"-\" без кавычек ").Return(tgbotapi.Message{}, nil).Once()
	mockTelegram.On("SendMessage", mock.AnythingOfType("int64"),
		"Отправьте фильтры разделённые пробелами. Если не хотите добавлять фильтры введите '-' без кавычек ").Return(tgbotapi.Message{}, nil).Once()
	mockTelegram.On("SendMessage", mock.AnythingOfType("int64"),
		"Ссылка отслеживается").Return(tgbotapi.Message{}, nil).Once()
	mockScrapper.On("AddLink", mock.AnythingOfType("int64"), domain.Link{URL: linkURL, Tags: []string{}, Filters: []string{}, ID: 0}).Return(nil)

	bot.Start()

	mockScrapper.AssertExpectations(t)
	mockTelegram.AssertExpectations(t)
}

func TestBot_Start_CommandUnTrack(t *testing.T) {
	mockTelegram := &mocks.TelegramClient{}
	mockScrapper := &mocks.ScrapperClient{}
	bot := application.NewBot(mockScrapper, mockTelegram, 8)
	command := "/untrack"
	linkURL := "https://github.com/TimofeyMosk/fractalFlame-image-creator"
	updates := make(chan tgbotapi.Update, 100)
	updates <- tgbotapi.Update{
		Message: &tgbotapi.Message{Text: command, Chat: &tgbotapi.Chat{ID: 123}, Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: len(command)}}},
	}
	updates <- tgbotapi.Update{
		Message: &tgbotapi.Message{Text: linkURL, Chat: &tgbotapi.Chat{ID: 123}},
	}
	updates <- tgbotapi.Update{
		Message: &tgbotapi.Message{Text: "-", Chat: &tgbotapi.Chat{ID: 123}},
	}
	updates <- tgbotapi.Update{
		Message: &tgbotapi.Message{Text: "-", Chat: &tgbotapi.Chat{ID: 123}},
	}
	close(updates)

	mockTelegram.On("GetUpdates").Return(tgbotapi.UpdatesChannel(updates)).Once()
	mockTelegram.On("SendMessage", mock.AnythingOfType("int64"),
		"Введите адрес ссылки").Return(tgbotapi.Message{}, nil).Once()
	mockTelegram.On("SendMessage", mock.AnythingOfType("int64"),
		"Ссылка успешно удалена").Return(tgbotapi.Message{}, nil).Once()

	mockScrapper.On("RemoveLink", mock.AnythingOfType("int64"), domain.Link{URL: linkURL}).Return(nil)

	bot.Start()

	mockScrapper.AssertExpectations(t)
	mockTelegram.AssertExpectations(t)
}

func TestBot_Start_CommandList(t *testing.T) {
	mockTelegram := &mocks.TelegramClient{}
	mockScrapper := &mocks.ScrapperClient{}
	bot := application.NewBot(mockScrapper, mockTelegram, 8)
	command := "/list"
	linkURL1 := "https://github.com/TimofeyMosk/fractalFlame-image-creator"
	linkURL2 := "https://github.com/central-university-dev/go-TimofeyMosk"
	updates := make(chan tgbotapi.Update, 100)
	updates <- tgbotapi.Update{
		Message: &tgbotapi.Message{Text: command, Chat: &tgbotapi.Chat{ID: 123}, Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: len(command)}}},
	}
	close(updates)

	mockTelegram.On("GetUpdates").Return(tgbotapi.UpdatesChannel(updates)).Once()
	mockTelegram.On("SendMessage", mock.AnythingOfType("int64"),
		"Список отслеживаемых ссылок:\nhttps://github.com/TimofeyMosk/fractalFlame-image-creator Tags: my  Filters: git \n"+
			"https://github.com/central-university-dev/go-TimofeyMosk Tags: AB  Filters: git \n").Return(tgbotapi.Message{}, nil).Once()

	mockScrapper.On("GetLinks", mock.AnythingOfType("int64")).Return([]domain.Link{
		{URL: linkURL1, Tags: []string{"my"}, Filters: []string{"git"}, ID: 0},
		{URL: linkURL2, Tags: []string{"AB"}, Filters: []string{"git"}, ID: 1}}, nil)

	bot.Start()

	mockScrapper.AssertExpectations(t)
	mockTelegram.AssertExpectations(t)
}
