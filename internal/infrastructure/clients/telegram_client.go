package clients

import (
	"fmt"
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramHTTPClient struct {
	bot     *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel
}

func NewTelegramHTTPClient(token string) (*TelegramHTTPClient, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	if err := setBotCommands(bot); err != nil {
		slog.Error("Failed to set bot commands: %v", "error", err.Error())
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	bot.Debug = false

	return &TelegramHTTPClient{
		bot:     bot,
		updates: updates,
	}, nil
}

func setBotCommands(bot *tgbotapi.BotAPI) error {
	cfg := tgbotapi.NewSetMyCommands(domain.BotCommands...)
	_, err := bot.Request(cfg)

	return err
}

func (c *TelegramHTTPClient) StopReceivingUpdates() {
	c.bot.StopReceivingUpdates()
}

func (c *TelegramHTTPClient) GetUpdates() tgbotapi.UpdatesChannel {
	return c.updates
}

func (c *TelegramHTTPClient) SendMessage(chatID int64, text string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	return c.bot.Send(msg)
}
