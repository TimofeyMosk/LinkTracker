package clients

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramAPIClient struct {
	bot     *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel
}

// NewTelegramAPI создаёт новый экземпляр telegramAPIClient.
func NewTelegramAPI(token string) (*TelegramAPIClient, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	return &TelegramAPIClient{
		bot:     bot,
		updates: updates,
	}, nil
}

// GetUpdates возвращает канал обновлений.
func (c *TelegramAPIClient) GetUpdates() tgbotapi.UpdatesChannel {
	return c.updates
}

// SendMessage отправляет сообщение и возвращает отправленное сообщение и ошибку.
func (c *TelegramAPIClient) SendMessage(chatID int64, text string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	return c.bot.Send(msg)
}
