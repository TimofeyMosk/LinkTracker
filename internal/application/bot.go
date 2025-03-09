package application

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	NotState = iota
	WaitingLink
	WaitingTags
	WaitingFilters
	WaitingDelete
)

type LinkWithState struct {
	Link  domain.Link
	state int
}

type ScrapperClient interface {
	RegisterUser(tgID int64) error
	DeleteUser(tgID int64) error
	AddLink(tgID int64, link domain.Link) error
	GetLinks(tgID int64) ([]domain.Link, error)
	RemoveLink(tgID int64, link domain.Link) error
}

type TelegramClient interface {
	GetUpdates() tgbotapi.UpdatesChannel
	SendMessage(chatID int64, text string) (tgbotapi.Message, error)
	StopReceivingUpdates()
}

type Bot struct {
	tgAPI              TelegramClient
	scrapperHTTPClient ScrapperClient
	userState          map[int64]LinkWithState
	mu                 sync.RWMutex
	workerCount        int
	globalLimiter      *rate.Limiter
}

func NewBot(scrapperClient ScrapperClient, tgClient TelegramClient, countWorkers int) *Bot {
	slog.Info("Bot create")

	return &Bot{
		tgAPI:              tgClient,
		scrapperHTTPClient: scrapperClient,
		userState:          make(map[int64]LinkWithState),
		workerCount:        countWorkers,
		globalLimiter:      rate.NewLimiter(rate.Every(time.Second/30), 30),
	}
}

func (bot *Bot) SendMessage(chatID int64, text string) {
	err := bot.globalLimiter.Wait(context.Background())
	if err != nil {
		slog.Error("Rate limit error", "error", err)
		return
	}

	message, err := bot.tgAPI.SendMessage(chatID, text)
	if err != nil {
		slog.Error("Failed to send message", "error", err.Error(), "chatId", chatID, "text", text)
		return
	}

	slog.Info("Sent message", "chatId", chatID, "text", message.Text)
}

func (bot *Bot) Start() {
	jobs := make([]chan *tgbotapi.Message, bot.workerCount)
	for i := range jobs {
		jobs[i] = make(chan *tgbotapi.Message, 100)
	}

	wg := sync.WaitGroup{}

	for i := 0; i < bot.workerCount; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			for msg := range jobs[workerID] {
				if msg.IsCommand() {
					bot.handleCommand(msg)
				} else {
					bot.changeState(msg)
				}
			}
		}(i)
	}

	updates := bot.tgAPI.GetUpdates()
	for update := range updates {
		if update.Message != nil {
			chatID := update.Message.Chat.ID
			index := chatID % int64(bot.workerCount)
			jobs[index] <- update.Message
		}
	}

	for i := range jobs {
		close(jobs[i])
	}

	wg.Wait()
	slog.Info("Bot has stopped")
}

func (bot *Bot) Stop() {
	slog.Info("Bot stopping. Waiting for running functions to execute")
	bot.tgAPI.StopReceivingUpdates()
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
	case "list":
		bot.commandList(message)
	default:
		bot.SendMessage(message.Chat.ID, "Команда не распознана. Введите /help , чтобы увидеть список доступных команд")
	}
}
func (bot *Bot) changeState(message *tgbotapi.Message) {
	if val, ok := bot.userState[message.Chat.ID]; ok {
		switch val.state {
		case NotState:
			return
		case WaitingLink:
			bot.stateWaitLink(message)
		case WaitingTags:
			bot.stateWaitTags(message)
		case WaitingFilters:
			bot.stateWaitFilters(message)
		case WaitingDelete:
			bot.stateWaitDelete(message)
		default:
			return
		}
	}
}

func (bot *Bot) commandStart(message *tgbotapi.Message) {
	slog.Info("Command /start execution", "chatId", message.Chat.ID)

	err := bot.scrapperHTTPClient.RegisterUser(message.Chat.ID)
	if err != nil {
		bot.SendMessage(message.Chat.ID, "Не удалось выполнить операцию")
		return
	}

	bot.SendMessage(message.Chat.ID, "Добро пожаловать в LinkTracker, "+
		"это приложение для отслеживание изменений на github и stackoverflow."+
		"Для получения списка команд введите /help")
}

func (bot *Bot) commandHelp(message *tgbotapi.Message) {
	slog.Info("Command /help execution", "chatId", message.Chat.ID)

	var sb strings.Builder

	sb.WriteString("📝Доступные команды:\n\n")

	for _, cmd := range domain.BotCommands {
		sb.WriteString(fmt.Sprintf("/%s - %s\n", cmd.Command, cmd.Description))
	}

	bot.SendMessage(message.Chat.ID, sb.String())
}

func (bot *Bot) commandTrack(message *tgbotapi.Message) {
	slog.Info("Command /track execution", "chatId", message.Chat.ID)
	bot.mu.Lock()
	bot.userState[message.Chat.ID] = LinkWithState{Link: domain.Link{}, state: WaitingLink}
	bot.mu.Unlock()
	bot.SendMessage(message.Chat.ID, "Введите адрес ссылки (поддерживается только gitHub и stackOverFlow")
}

func (bot *Bot) commandUntrack(message *tgbotapi.Message) {
	slog.Info("Command /untrack execution", "chatId", message.Chat.ID)
	bot.mu.Lock()
	bot.userState[message.Chat.ID] = LinkWithState{Link: domain.Link{}, state: WaitingDelete}
	bot.mu.Unlock()
	bot.SendMessage(message.Chat.ID, "Введите адрес ссылки")
}

func (bot *Bot) commandList(message *tgbotapi.Message) {
	slog.Info("Command /list execution", "chatId", message.Chat.ID)

	list, err := bot.scrapperHTTPClient.GetLinks(message.Chat.ID)
	if err != nil {
		slog.Error("Failed to get links", "error", err.Error(), "chatId", message.Chat.ID)
		bot.SendMessage(message.Chat.ID, "Не удалось выполнить операцию")

		return
	}

	text := listLinksToString(list)

	bot.SendMessage(message.Chat.ID, text)
}

func (bot *Bot) stateWaitLink(message *tgbotapi.Message) {
	link := message.Text
	if validateLink(link) {
		bot.mu.Lock()
		bot.userState[message.Chat.ID] = LinkWithState{
			Link:  domain.Link{URL: link},
			state: WaitingTags}
		bot.mu.Unlock()
		bot.SendMessage(message.Chat.ID, "Отправьте теги разделённые пробелами. Если не хотите добавлять теги введите \"-\" без кавычек ")
	} else {
		bot.SendMessage(message.Chat.ID, "Поддерживается только gitHub(https://github.com/{owner}/{repo}) и "+
			"stackOverflow(https://stackoverflow.com/questions/{id}). Повторите команду /track")
		bot.mu.Lock()
		delete(bot.userState, message.Chat.ID)
		bot.mu.Unlock()
	}
}

func (bot *Bot) stateWaitTags(message *tgbotapi.Message) {
	bot.mu.RLock()
	tempLinkWithState := bot.userState[message.Chat.ID]
	bot.mu.RUnlock()

	if message.Text == "-" {
		tempLinkWithState.Link.Tags = []string{}
	} else {
		tempLinkWithState.Link.Tags = strings.Split(message.Text, " ")
	}

	tempLinkWithState.state = WaitingFilters

	bot.mu.Lock()
	bot.userState[message.Chat.ID] = tempLinkWithState
	bot.mu.Unlock()
	bot.SendMessage(message.Chat.ID, "Отправьте фильтры разделённые пробелами. Если не хотите добавлять фильтры введите '-' без кавычек ")
}

func (bot *Bot) stateWaitFilters(message *tgbotapi.Message) {
	bot.mu.RLock()
	tempLinkWithState := bot.userState[message.Chat.ID]
	bot.mu.RUnlock()

	if message.Text == "-" {
		tempLinkWithState.Link.Filters = []string{}
	} else {
		tempLinkWithState.Link.Filters = strings.Split(message.Text, " ")
	}

	tempLinkWithState.state = NotState

	err := bot.scrapperHTTPClient.AddLink(message.Chat.ID, tempLinkWithState.Link)
	if err != nil {
		slog.Error(err.Error())
		bot.SendMessage(message.Chat.ID, "Не удалось выполнить операцию")

		return
	}

	bot.SendMessage(message.Chat.ID, "Ссылка отслеживается")
	bot.mu.Lock()
	delete(bot.userState, message.Chat.ID)
	bot.mu.Unlock()
}

func (bot *Bot) stateWaitDelete(message *tgbotapi.Message) {
	link := message.Text
	if validateLink(link) {
		err := bot.scrapperHTTPClient.RemoveLink(message.Chat.ID, domain.Link{URL: link})
		if err != nil {
			slog.Error(err.Error())
			return
		}

		bot.SendMessage(message.Chat.ID, "Ссылка успешно удалена")
	} else {
		bot.SendMessage(message.Chat.ID, "Поддерживается только gitHub и stackOverflow. Не удалось выполнить операцию")
	}

	bot.mu.Lock()
	delete(bot.userState, message.Chat.ID)
	bot.mu.Unlock()
}

func validateLink(link string) bool {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return false
	}

	const (
		github        = "github.com"
		stackoverflow = "stackoverflow.com"
	)

	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")

	if parsedURL.Host == github && len(parts) == 2 {
		return true
	}

	if parsedURL.Host == stackoverflow && len(parts) >= 2 && len(parts) <= 3 && parts[0] == "questions" {
		return true
	}

	return false
}

func listLinksToString(links []domain.Link) string {
	var sb strings.Builder

	sb.WriteString("Список отслеживаемых ссылок:\n")

	for i := range links {
		sb.WriteString(links[i].URL)

		if len(links[i].Tags) > 0 {
			sb.WriteString(" Tags: ")

			for j := range links[i].Tags {
				sb.WriteString(links[i].Tags[j] + " ")
			}
		}

		if len(links[i].Filters) > 0 {
			sb.WriteString(" Filters: ")

			for j := range links[i].Filters {
				sb.WriteString(links[i].Filters[j] + " ")
			}
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
