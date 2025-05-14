package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"unicode/utf8"

	"LinkTracker/internal/domain"
)

const (
	WaitingLink = iota
	WaitingTags
	WaitingFilters
	WaitingDelete
	WaitingSetTagsWaitingLink
	WaitingSetTagsWaitingTags
)

const errorText = "Не удалось выполнить операцию"

type StateManager interface {
	CreateState(ctx context.Context, tgID int64, state int) error
	DeleteState(ctx context.Context, tgID int64) error
	GetState(ctx context.Context, tgID int64) (state int, link domain.Link, err error)
	UpdateState(ctx context.Context, tgID int64, state int, link *domain.Link) error
}

type ScrapperClient interface {
	RegisterUser(ctx context.Context, tgID int64) error
	DeleteUser(ctx context.Context, tgID int64) error
	AddLink(ctx context.Context, tgID int64, link *domain.Link) error
	GetLinks(ctx context.Context, tgID int64) ([]domain.Link, error)
	RemoveLink(ctx context.Context, tgID int64, link *domain.Link) error
	UpdateLink(ctx context.Context, tgID int64, link *domain.Link) error
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
	slog.Info("Bot create")

	return &Bot{
		scrapper: scrapperClient,
		tgAPI:    tgAPI,
	}
}

func (bot *Bot) UpdateSend(ctx context.Context, tgIDs []int64, linkURL, description string) {
	for _, tgID := range tgIDs {
		message := fmt.Sprintf("Было обновление: %s\n%s", linkURL, description)
		bot.tgAPI.SendMessage(ctx, tgID, message)
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

	slog.Info("Bot running", "workerCount", workerCount)
	<-ctx.Done()
	bot.tgAPI.StopReceiveMessage()
	close(messageChannel)

	for i := range jobs {
		close(jobs[i])
	}

	wg.Wait()
	slog.Info("Bot shutdown")
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
	case "/settags":
		return bot.commandSetTags(ctx, tgID)
	default:
		responseText := "Команда не распознана. Введите /help , чтобы увидеть список доступных команд"
		return responseText
	}
}

func (bot *Bot) changeState(ctx context.Context, tgID int64, text string) string {
	state, link, err := bot.scrapper.GetState(ctx, tgID)
	if err != nil {
		return ""
	}

	switch state {
	case WaitingLink:
		return bot.stateWaitLink(ctx, tgID, text, &link)
	case WaitingTags:
		return bot.stateWaitTags(ctx, tgID, text, &link)
	case WaitingFilters:
		return bot.stateWaitFilters(ctx, tgID, text, &link)
	case WaitingDelete:
		return bot.stateWaitDelete(ctx, tgID, text)
	case WaitingSetTagsWaitingLink:
		return bot.stateSetTagsWaitingLink(ctx, tgID, text)
	case WaitingSetTagsWaitingTags:
		return bot.stateSetTagsWaitingTags(ctx, tgID, text, &link)
	default:
		return ""
	}
}

func (bot *Bot) commandStart(ctx context.Context, id int64) string {
	err := bot.scrapper.RegisterUser(ctx, id)
	if err != nil {
		slog.Error("Command /start failed", "error", err, "chatId", id)
		return errorText + ". Возможно, вы уже зарегистрированы в приложении"
	}

	responseText := "Добро пожаловать в LinkTracker, " +
		"это приложение для отслеживание изменений на github и stackoverflow." +
		"Для получения списка команд введите /help"

	slog.Info("Command /start done", "chatId", id)

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
	err := bot.scrapper.CreateState(ctx, tgID, WaitingLink)
	if err != nil {
		slog.Error("Command /track failed", "error", err, "chatId", tgID)
		return errorText
	}

	responseText := "Введите адрес ссылки (поддерживается только gitHub и stackOverFlow)"

	slog.Info("Command /track done", "chatId", tgID)

	return responseText
}

func (bot *Bot) commandUntrack(ctx context.Context, tgID int64) string {
	slog.Info("Command /untrack execution", "chatId", tgID)

	err := bot.scrapper.CreateState(ctx, tgID, WaitingDelete)
	if err != nil {
		return errorText
	}

	responseText := "Введите адрес ссылки для удаления"

	return responseText
}

func (bot *Bot) commandSetTags(ctx context.Context, tgID int64) string {
	slog.Info("Command /settags execution", "chatId", tgID)

	err := bot.scrapper.CreateState(ctx, tgID, WaitingSetTagsWaitingLink)
	if err != nil {
		return errorText
	}

	responseText := "Введите ссылку, для которой хотите изменить тег/теги"

	return responseText
}

func (bot *Bot) commandList(ctx context.Context, tgID int64) string {
	list, err := bot.scrapper.GetLinks(ctx, tgID)
	if err != nil {
		slog.Error("Command /list failed", "error", err.Error(), "chatId", tgID)

		return errorText
	}

	slog.Info("Command /list done", "chatId", tgID)

	if len(list) == 0 {
		responseText := "Список отслеживаемых ссылок пуст. Добавьте ссылки с помощью /track"
		return responseText
	}

	responseText := listLinksToString(list)

	return responseText
}

func (bot *Bot) stateWaitLink(ctx context.Context, tgID int64, text string, link *domain.Link) string {
	linkURL := text
	valid, validURL := validateLink(linkURL)

	if !valid {
		err := bot.scrapper.DeleteState(ctx, tgID)
		if err != nil {
			slog.Error("stateWaitLink failed", "error", err.Error(), "chatId", tgID)
			return errorText
		}

		slog.Info("stateWaitLink  done", "chatId", tgID)

		responseText := "Поддерживается только gitHub(https://github.com/{owner}/{repo}) и " +
			"stackOverflow(https://stackoverflow.com/questions/{id}). Повторите команду /track"

		return responseText
	}

	link.URL = validURL

	err := bot.scrapper.UpdateState(ctx, tgID, WaitingTags, link)
	if err != nil {
		slog.Error("stateWaitLink failed", "error", err.Error(), "chatId", tgID)
		return errorText
	}

	slog.Info("stateWaitLink  done", "chatId", tgID)

	responseText := "Отправьте теги разделённые пробелами. Если не хотите добавлять теги введите \"-\" без кавычек"

	return responseText
}

func (bot *Bot) stateWaitTags(ctx context.Context, tgID int64, text string, link *domain.Link) string {
	if text == "-" {
		link.Tags = []string{}

		err := bot.scrapper.UpdateState(ctx, tgID, WaitingFilters, link)
		if err != nil {
			slog.Error("stateWaitTags failed", "error", err.Error(), "chatId", tgID)
			return errorText
		}
	} else {
		link.Tags = strings.Split(text, " ")

		err := bot.scrapper.UpdateState(ctx, tgID, WaitingFilters, link)
		if err != nil {
			slog.Error("stateWaitTags failed", "error", err.Error(), "chatId", tgID)
			return errorText
		}
	}

	slog.Info("stateWaitTags  done", "chatId", tgID)

	responseText := "Отправьте фильтры разделённые пробелами. Если не хотите добавлять фильтры введите '-' без кавычек"

	return responseText
}

func (bot *Bot) stateWaitFilters(ctx context.Context, tgID int64, text string, link *domain.Link) string {
	if text == "-" {
		link.Filters = []string{}
	} else {
		link.Filters = strings.Split(text, " ")
	}

	err := bot.scrapper.AddLink(ctx, tgID, link)
	if err != nil {
		if errors.As(err, &domain.ErrAPI{}) && (err.(domain.ErrAPI).ExceptionMessage == domain.ErrLinkAlreadyTracking{}.Error()) {
			err = bot.scrapper.DeleteState(ctx, tgID)
			if err != nil {
				return errorText
			}

			slog.Info("stateWaitTags  done", "chatId", tgID)

			responseText := "Данная ссылка уже отслеживается"

			return responseText
		}

		slog.Error("stateWaitFilters failed", "error", err.Error(), "chatId", tgID)

		return errorText
	}

	err = bot.scrapper.DeleteState(ctx, tgID)
	if err != nil {
		slog.Error("stateWaitFilters failed", "error", err.Error(), "chatId", tgID)
		return errorText
	}

	slog.Info("stateWaitTags  done", "chatId", tgID)

	responseText := "Ссылка отслеживается"

	return responseText
}

func (bot *Bot) stateWaitDelete(ctx context.Context, tgID int64, text string) string {
	link := text

	err := bot.scrapper.RemoveLink(ctx, tgID, &domain.Link{URL: link})
	if err != nil {
		slog.Error("stateWaitDelete failed", "error", err.Error(), "chatId", tgID)
		return errorText
	}

	err = bot.scrapper.DeleteState(ctx, tgID)
	if err != nil {
		slog.Error("stateWaitDelete failed", "error", err.Error(), "chatId", tgID)
		return errorText
	}

	slog.Info("stateWaitDelete done", "chatId", tgID)

	responseText := "Ссылка успешно удалена"

	return responseText
}

func (bot *Bot) stateSetTagsWaitingLink(ctx context.Context, tgID int64, text string) string {
	links, err := bot.scrapper.GetLinks(ctx, tgID)
	if err != nil {
		slog.Error("stateSetTagsWaitingLink failed", "error", err.Error(), "chatId", tgID)
		return errorText
	}

	var link *domain.Link

	for i := range links {
		if links[i].URL == text {
			link = &links[i]
		}
	}

	if link == nil {
		slog.Error("stateSetTagsWaitingLink failed", "error", "Данная ссылка не найдена", "chatId", tgID)
		return errorText + ". Данная ссылка не найдена"
	}

	err = bot.scrapper.UpdateState(ctx, tgID, WaitingSetTagsWaitingTags, link)
	if err != nil {
		slog.Error("stateSetTagsWaitingLink failed", "error", err.Error(), "chatId", tgID)
		return errorText
	}

	slog.Info("stateSetTagsWaitingLink done", "chatId", tgID)

	responseText := "Отправьте новые теги разделённые пробелами. Если не хотите добавлять теги отправьте '-' без кавычек"

	return responseText
}

func (bot *Bot) stateSetTagsWaitingTags(ctx context.Context, tgID int64, text string, link *domain.Link) string {
	if text == "-" {
		link.Tags = []string{}
	} else {
		link.Tags = strings.Split(text, " ")
	}

	err := bot.scrapper.UpdateLink(ctx, tgID, link)
	if err != nil {
		slog.Error("stateSetTagsWaitingTags done", "error", err.Error(), "chatId", tgID)
		return errorText
	}

	slog.Info("stateSetTagsWaitingTags done", "chatId", tgID)

	responseText := "Теги успешно изменены"

	return responseText
}

func validateLink(link string) (valid bool, validURL string) {
	parsedURL, err := url.Parse(link)
	if err != nil {
		slog.Error("validateLink failed", "error", err.Error(), "link", link)
		return false, ""
	}

	const (
		github        = "github.com"
		stackoverflow = "stackoverflow.com"
	)

	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")

	if parsedURL.Host == github && len(parts) >= 2 {
		validURL, err := url.JoinPath(parsedURL.Scheme+"://"+parsedURL.Host, parts[0], parts[1])
		if err != nil {
			slog.Error("validateLink failed", "error", err.Error(), "link", link)
			return false, ""
		}

		return true, validURL
	}

	if parsedURL.Host == stackoverflow && len(parts) >= 2 && parts[0] == "questions" {
		validURL, err := url.JoinPath(parsedURL.Scheme+"://"+parsedURL.Host, parts[0], parts[1])
		if err != nil {
			slog.Error("validateLink failed", "error", err.Error(), "link", link)
			return false, ""
		}

		return true, validURL
	}

	return false, ""
}

func formatLink(link *domain.Link) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("linkID: %d Url: %s", link.ID, link.URL))

	if len(link.Tags) > 0 {
		parts = append(parts, "Tags: "+strings.Join(link.Tags, " "))
	}

	if len(link.Filters) > 0 {
		parts = append(parts, "Filters: "+strings.Join(link.Filters, " "))
	}

	return strings.Join(parts, " ")
}

func listLinksToString(links []domain.Link) string {
	tagsLinks := make(map[string][]domain.Link)

	for _, link := range links {
		for _, tag := range link.Tags {
			tagsLinks[tag] = append(tagsLinks[tag], link)
		}
	}

	var sb strings.Builder

	sb.WriteString("Список отслеживаемых ссылок:\n")

	if len(tagsLinks) > 0 {
		for tag, taggedlinks := range tagsLinks {
			sb.WriteString(tag + ": \n")

			for _, link := range taggedlinks {
				sb.WriteString(formatLink(&link) + "\n")
			}

			sb.WriteString("\n")
		}
	}

	sb.WriteString("Без тегов: \n")

	for i := range links {
		if len(links[i].Tags) == 0 {
			sb.WriteString(formatLink(&links[i]))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
