package clients

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/time/rate"
)

type Bot interface {
	HandleMessage(id int64, text string) string
}

type TelegramHTTPClient struct {
	tgBotAPI      *tgbotapi.BotAPI
	bot           Bot
	updates       tgbotapi.UpdatesChannel
	globalLimiter *rate.Limiter
}

func NewTelegramHTTPClient(token string, bot Bot) (*TelegramHTTPClient, error) {
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
		bot:           bot,
		updates:       updates,
		globalLimiter: rate.NewLimiter(rate.Every(time.Second/30), 30),
	}, nil
}

func (t *TelegramHTTPClient) Run() {
	workerCount := 17

	jobs := make([]chan *tgbotapi.Message, workerCount)
	for i := range jobs {
		jobs[i] = make(chan *tgbotapi.Message, 100)
	}

	wg := sync.WaitGroup{}
	for i := 0; i < workerCount; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			for msg := range jobs[workerID] {
				chatID := msg.Chat.ID
				messageText := msg.Text

				responseText := t.bot.HandleMessage(chatID, messageText)
				if responseText != "" {
					t.SendMessage(chatID, responseText)
				}
			}
		}(i)
	}

	for update := range t.updates {
		if update.Message != nil {
			chatID := update.Message.Chat.ID
			numWorker := chatID % int64(workerCount)
			jobs[numWorker] <- update.Message
		}
	}

	for i := range jobs {
		close(jobs[i])
	}

	wg.Wait()
	slog.Info("Telegram client has finished")
}

func (t *TelegramHTTPClient) Stop() {
	t.tgBotAPI.StopReceivingUpdates()
}

func setBotCommands(bot *tgbotapi.BotAPI, botCommands []tgbotapi.BotCommand) error {
	cfg := tgbotapi.NewSetMyCommands(botCommands...)
	_, err := bot.Request(cfg)

	return err
}

func (t *TelegramHTTPClient) SendMessage(chatID int64, text string) {
	err := t.globalLimiter.Wait(context.Background())
	if err != nil {
		slog.Error("Rate limit error", "chatID", chatID, "text", text, "error", err.Error())
	}

	msg := tgbotapi.NewMessage(chatID, text)

	_, err = t.tgBotAPI.Send(msg)
	if err != nil {
		slog.Error(err.Error())
	}
}
