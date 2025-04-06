package bot

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"LinkTracker/internal/domain"
)

const (
	NotState = iota
	WaitingLink
	WaitingTags
	WaitingFilters
	WaitingDelete
	WaitingReplaceTags
)

const errorText = "Не удалось выполнить операцию"

type StateManager interface {
	CreateState(ctx context.Context, tgID int64, state int) error
	DeleteState(ctx context.Context, tgID int64) error
	GetState(ctx context.Context, tgID int64) (state int, err error)
	UpdateState(ctx context.Context, tgID int64, state int) error
	UpdateURL(ctx context.Context, tgID int64, linkURL string) error
	UpdateTags(ctx context.Context, tgID int64, tags []string) error
	UpdateFilters(ctx context.Context, tgID int64, filters []string) error
	GetStateLink(ctx context.Context, tgID int64) (link domain.Link, err error)
}

type ScrapperClient interface {
	RegisterUser(ctx context.Context, tgID int64) error
	DeleteUser(ctx context.Context, tgID int64) error
	AddLink(ctx context.Context, tgID int64, link *domain.Link) error
	GetLinks(ctx context.Context, tgID int64) ([]domain.Link, error)
	RemoveLink(ctx context.Context, tgID int64, link *domain.Link) error
	StateManager
}

type TelegramClient interface {
	SendMessage(ctx context.Context, tgID int64, text string)
	ReceiveMessage(messageCh chan domain.Message)
	StopReceiveMessage()
}

type Bot struct {
	scrapper ScrapperClient
	tgAPI    TelegramClient
}

func NewBot(scrapperClient ScrapperClient, tgAPI TelegramClient) *Bot {
	slog.Info("tgBOTApi create")

	return &Bot{
		scrapper: scrapperClient,
		tgAPI:    tgAPI,
	}
}

func (bot *Bot) Run(ctx context.Context) {
	workerCount := 17

	jobs := make([]chan domain.Message, workerCount)
	for i := range jobs {
		jobs[i] = make(chan domain.Message, 100)
	}

	wg := sync.WaitGroup{}
	for i := 0; i < workerCount; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			for msg := range jobs[workerID] {
				responseText := bot.HandleMessage(ctx, msg.TgID, msg.Text)
				if responseText != "" {
					bot.tgAPI.SendMessage(ctx, msg.TgID, responseText)
				}
			}
		}(i)
	}

	messageChannel := make(chan domain.Message, 100)

	go bot.tgAPI.ReceiveMessage(messageChannel)

	wg.Add(1)

	go func() {
		defer wg.Done()

		for message := range messageChannel {
			numWorker := message.TgID % int64(workerCount)
			jobs[numWorker] <- message
		}
	}()

	<-ctx.Done()
	bot.tgAPI.StopReceiveMessage()
	close(messageChannel)

	for i := range jobs {
		close(jobs[i])
	}

	wg.Wait()
	slog.Info("bot shutdown")
}

func (bot *Bot) HandleMessage(ctx context.Context, id int64, text string) string {
	if firstRune, _ := utf8.DecodeRuneInString(text); firstRune == '/' {
		return bot.handleCommand(ctx, id, text)
	}

	return bot.changeState(ctx, id, text)
}

func (bot *Bot) handleCommand(ctx context.Context, tgID int64, text string) string {
	switch text {
	case "/start":
		return bot.commandStart(ctx, tgID)
	case "/help":
		return bot.commandHelp(tgID)
	case "/track":
		return bot.commandTrack(ctx, tgID)
	case "/untrack":
		return bot.commandUntrack(ctx, tgID)
	case "/list":
		return bot.commandList(ctx, tgID)
	default:
		responseText := "Команда не распознана. Введите /help , чтобы увидеть список доступных команд"
		return responseText
	}
}

func (bot *Bot) changeState(ctx context.Context, tgID int64, text string) string {
	state, err := bot.scrapper.GetState(ctx, tgID)
	if err != nil {
		return errorText
	}

	switch state {
	case NotState:
		return ""
	case WaitingLink:
		return bot.stateWaitLink(ctx, tgID, text)
	case WaitingTags:
		return bot.stateWaitTags(ctx, tgID, text)
	case WaitingFilters:
		return bot.stateWaitFilters(ctx, tgID, text)
	case WaitingDelete:
		return bot.stateWaitDelete(ctx, tgID, text)
	default:
		return ""
	}
}

func (bot *Bot) commandStart(ctx context.Context, id int64) string {
	slog.Info("Command /start execution", "chatId", id)

	err := bot.scrapper.RegisterUser(ctx, id)
	if err != nil {
		return errorText
	}

	responseText := "Добро пожаловать в LinkTracker, " +
		"это приложение для отслеживание изменений на github и stackoverflow." +
		"Для получения списка команд введите /help"

	return responseText
}

func (bot *Bot) commandHelp(id int64) string {
	slog.Info("Command /help execution", "chatId", id)

	responseText := "📝Доступные команды:\n\n" +
		"/start - Начало работы с ботом\n" +
		"/help - Помощь по командам\n" +
		"/track - Начать отслеживание ссылки\n" +
		"/untrack - Прекратить отслеживание\n" +
		"/list - Список отслеживаемых ссылок"

	return responseText
}

func (bot *Bot) commandTrack(ctx context.Context, tgID int64) string {
	slog.Info("Command /track execution", "chatId", tgID)

	err := bot.scrapper.UpdateState(ctx, tgID, WaitingLink)
	if err != nil {
		return errorText
	}

	responseText := "Введите адрес ссылки (поддерживается только gitHub и stackOverFlow)"

	return responseText
}

func (bot *Bot) commandUntrack(ctx context.Context, tgID int64) string {
	slog.Info("Command /untrack execution", "chatId", tgID)

	err := bot.scrapper.UpdateState(ctx, tgID, WaitingDelete)
	if err != nil {
		return errorText
	}

	responseText := "Введите адрес ссылки для удаления"

	return responseText
}

func (bot *Bot) commandList(ctx context.Context, tgID int64) string {
	slog.Info("Command /list execution", "chatId", tgID)

	list, err := bot.scrapper.GetLinks(ctx, tgID)
	if err != nil {
		slog.Error("Failed to get links", "error", err.Error(), "chatId", tgID)

		return errorText
	}

	if len(list) == 0 {
		responseText := "Список отслеживаемых ссылок пуст. Добавьте ссылки с помощью /track"
		return responseText
	}

	responseText := listLinksToString(list)

	return responseText
}

func (bot *Bot) stateWaitLink(ctx context.Context, tgID int64, messageURL string) string {
	link := messageURL
	if !validateLink(link) {

		err := bot.scrapper.DeleteState(ctx, tgID)
		if err != nil {
			return errorText
		}

		responseText := "Поддерживается только gitHub(https://github.com/{owner}/{repo}) и " +
			"stackOverflow(https://stackoverflow.com/questions/{id}). Повторите команду /track"

		return responseText
	}

	err := bot.scrapper.UpdateURL(ctx, tgID, link)
	if err != nil {
		return errorText
	}

	err = bot.scrapper.UpdateState(ctx, tgID, WaitingTags)
	if err != nil {
		return errorText
	}

	responseText := "Отправьте теги разделённые пробелами. Если не хотите добавлять теги введите \"-\" без кавычек"

	return responseText
}

func (bot *Bot) stateWaitTags(ctx context.Context, tgID int64, text string) string {
	if text == "-" {
		err := bot.scrapper.UpdateTags(ctx, tgID, []string{})
		if err != nil {
			return errorText
		}
	} else {
		tags := strings.Split(text, " ")
		err := bot.scrapper.UpdateTags(ctx, tgID, tags)
		if err != nil {
			return errorText
		}
	}

	err := bot.scrapper.UpdateState(ctx, tgID, WaitingFilters)
	if err != nil {
		return errorText
	}

	responseText := "Отправьте фильтры разделённые пробелами. Если не хотите добавлять фильтры введите '-' без кавычек"

	return responseText
}

func (bot *Bot) stateWaitFilters(ctx context.Context, tgID int64, text string) string {

	if text == "-" {
		err := bot.scrapper.UpdateFilters(ctx, tgID, []string{})
		if err != nil {
			return errorText
		}
	} else {
		filters := strings.Split(text, " ")
		err := bot.scrapper.UpdateFilters(ctx, tgID, filters)
		if err != nil {
			return errorText
		}
	}

	link, err := bot.scrapper.GetStateLink(ctx, tgID)
	if err != nil {
		return errorText
	}

	err = bot.scrapper.AddLink(ctx, tgID, &link)
	if err != nil {
		slog.Error(err.Error())

		if errors.As(err, &domain.ErrAPI{}) && (err.(domain.ErrAPI).ExceptionMessage == domain.ErrLinkAlreadyTracking{}.Error()) {
			responseText := "Данная ссылка уже отслеживается"
			return responseText
		}

		return errorText
	}

	err = bot.scrapper.DeleteState(ctx, tgID)
	if err != nil {
		return errorText
	}

	responseText := "Ссылка отслеживается"

	return responseText
}

func (bot *Bot) stateWaitDelete(ctx context.Context, id int64, text string) string {
	link := text

	err := bot.scrapper.RemoveLink(ctx, id, &domain.Link{URL: link})
	if err != nil {
		slog.Error(err.Error())

		return errorText
	}

	err = bot.scrapper.DeleteState(ctx, id)
	if err != nil {
		return errorText
	}

	responseText := "Ссылка успешно удалена"

	return responseText
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

	if parsedURL.Host == github && len(parts) >= 2 {
		return true
	}

	if parsedURL.Host == stackoverflow && len(parts) >= 2 && parts[0] == "questions" {
		return true
	}

	return false
}

func listLinksToString(links []domain.Link) string {
	var sb strings.Builder

	sb.WriteString("Список отслеживаемых ссылок:\n")

	for i := range links {
		sb.WriteString("linkID: " + strconv.Itoa(int(links[i].ID)))
		sb.WriteString(" Url: " + links[i].URL)

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
