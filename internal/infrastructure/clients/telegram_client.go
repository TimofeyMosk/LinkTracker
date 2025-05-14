package clients

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"LinkTracker/internal/domain"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/time/rate"
)

type TelegramHTTPClient struct {
	tgBotAPI      *tgbotapi.BotAPI
	updates       tgbotapi.UpdatesChannel
	globalLimiter *rate.Limiter
}

func NewTelegramHTTPClient(token string) (*TelegramHTTPClient, error) {
	tgBotAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	botCommands := []tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "Начало работы с ботом",
		},
		{
			Command:     "help",
			Description: "Помощь по командам",
		},
		{
			Command:     "track",
			Description: "Начать отслеживание ссылки",
		},
		{
			Command:     "untrack",
			Description: "Прекратить отслеживание",
		},
		{
			Command:     "settags",
			Description: "Изменить теги у ссылки",
		},
		{
			Command:     "list",
			Description: "Список отслеживаемых ссылок",
		},
	}

	if err := setBotCommands(tgBotAPI, botCommands); err != nil {
		slog.Error("Failed to set bot commands: %v", "error", err.Error())
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := tgBotAPI.GetUpdatesChan(u)
	tgBotAPI.Debug = false

	return &TelegramHTTPClient{
		tgBotAPI:      tgBotAPI,
		updates:       updates,
		globalLimiter: rate.NewLimiter(rate.Every(time.Second/30), 30),
	}, nil
}

func (t *TelegramHTTPClient) ReceiveMessage(messageCh chan domain.Message) {
	for update := range t.updates {
		if update.Message != nil {
			messageCh <- domain.Message{TgID: update.Message.From.ID, Text: update.Message.Text}
		}
	}
}

func (t *TelegramHTTPClient) StopReceiveMessage() {
	t.tgBotAPI.StopReceivingUpdates()
}

func (t *TelegramHTTPClient) SendMessage(ctx context.Context, chatID int64, text string) {
	err := t.globalLimiter.Wait(ctx)
	if err != nil {
		slog.Error("Rate limit error", "chatID", chatID, "text", text, "error", err.Error())
		return
	}

	msg := tgbotapi.NewMessage(chatID, text)

	_, err = t.tgBotAPI.Send(msg)
	if err != nil {
		slog.Error(err.Error())
	}
}

func setBotCommands(bot *tgbotapi.BotAPI, botCommands []tgbotapi.BotCommand) error {
	cfg := tgbotapi.NewSetMyCommands(botCommands...)
	_, err := bot.Request(cfg)

	return err
}
