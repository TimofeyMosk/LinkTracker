package application

import (
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients"
)

type BotHTTPClient interface {
	RegisterUser(tgID int64) error
}

type Bot struct {
	tgAPI         *clients.TelegramAPIClient
	botHTTPClient BotHTTPClient
}

func NewBot(tgToken string, botClient BotHTTPClient) (*Bot, error) {
	tgAPI, err := clients.NewTelegramAPI(tgToken)
	if err != nil {
		return nil, err
	}

	return &Bot{
		tgAPI:         tgAPI,
		botHTTPClient: botClient,
	}, nil
}

func (bot *Bot) Start() {
	updates := bot.tgAPI.GetUpdates()

	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				bot.handleCommand(update.Message)
			} else {
				// Обработка других типов сообщений
			}
		}
	}
}

func (bot *Bot) handleCommand(message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		bot.commandStart(message)
	case "help":
		bot.commandHelp(message)
	case "track":
		bot.commandTrack(message)
	case "untrack":
		bot.commandUntrack(message)
	default:
		// Логика для неизвестных команд
	}
}

func (bot *Bot) commandStart(message *tgbotapi.Message) {
	err := bot.botHTTPClient.RegisterUser(message.Chat.ID)
	if err != nil {
		slog.Error(err.Error())

		_, err = bot.tgAPI.SendMessage(message.Chat.ID, "Не удалось выполнить операцию. Повторите попытку позже")
		if err != nil {
			slog.Error(err.Error())
		}

		return
	}

	_, err = bot.tgAPI.SendMessage(message.Chat.ID, "Добро пожаловать в LinkTracker.Для получения списка команд введите /help")
	if err != nil {
		slog.Error(err.Error())
	}
}

func (bot *Bot) commandHelp(message *tgbotapi.Message) {
	text := `/start - регистрация пользователя.
/help - вывод списка доступных команд.
/track - начать отслеживание ссылки.
/untrack - прекратить отслеживание ссылки.
/list - показать список отслеживаемых ссылок (cписок ссылок, отправленных при /track)
`

	_, err := bot.tgAPI.SendMessage(message.Chat.ID, text)
	if err != nil {
		slog.Error(err.Error())
	}
}

func (bot *Bot) commandTrack(message *tgbotapi.Message) {

}

func (bot *Bot) commandUntrack(message *tgbotapi.Message) {

}
