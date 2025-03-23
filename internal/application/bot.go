package application

import (
	"errors"
	"log/slog"
	"net/url"
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
)
const errorText = "Не удалось выполнить операцию"

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

type Bot struct {
	scrapperHTTPClient ScrapperClient
	userState          map[int64]LinkWithState
	mu                 sync.RWMutex
}

func NewBot(scrapperClient ScrapperClient) *Bot {
	slog.Info("tgBOTApi create")

	return &Bot{
		scrapperHTTPClient: scrapperClient,
		userState:          make(map[int64]LinkWithState),
	}
}

func (bot *Bot) HandleMessage(id int64, text string) string {
	if firstRune, _ := utf8.DecodeRuneInString(text); firstRune == '/' {
		return bot.handleCommand(id, text)
	}

	return bot.changeState(id, text)
}

func (bot *Bot) handleCommand(id int64, text string) string {
	switch text {
	case "/start":
		return bot.commandStart(id)
	case "/help":
		return bot.commandHelp(id)
	case "/track":
		return bot.commandTrack(id)
	case "/untrack":
		return bot.commandUntrack(id)
	case "/list":
		return bot.commandList(id)
	default:
		responseText := "Команда не распознана. Введите /help , чтобы увидеть список доступных команд"
		return responseText
	}
}

func (bot *Bot) changeState(id int64, text string) string {
	if val, ok := bot.userState[id]; ok {
		switch val.state {
		case NotState:
			return ""
		case WaitingLink:
			return bot.stateWaitLink(id, text)
		case WaitingTags:
			return bot.stateWaitTags(id, text)
		case WaitingFilters:
			return bot.stateWaitFilters(id, text)
		case WaitingDelete:
			return bot.stateWaitDelete(id, text)
		default:
			return ""
		}
	}

	return ""
}

func (bot *Bot) commandStart(id int64) string {
	slog.Info("Command /start execution", "chatId", id)

	err := bot.scrapperHTTPClient.RegisterUser(id)
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

func (bot *Bot) commandTrack(id int64) string {
	slog.Info("Command /track execution", "chatId", id)
	bot.mu.Lock()
	bot.userState[id] = LinkWithState{Link: domain.Link{}, state: WaitingLink}
	bot.mu.Unlock()

	responseText := "Введите адрес ссылки (поддерживается только gitHub и stackOverFlow)"

	return responseText
}

func (bot *Bot) commandUntrack(id int64) string {
	slog.Info("Command /untrack execution", "chatId", id)
	bot.mu.Lock()
	bot.userState[id] = LinkWithState{Link: domain.Link{}, state: WaitingDelete}
	bot.mu.Unlock()

	responseText := "Введите адрес ссылки для удаления"

	return responseText
}

func (bot *Bot) commandList(id int64) string {
	slog.Info("Command /list execution", "chatId", id)

	list, err := bot.scrapperHTTPClient.GetLinks(id)
	if err != nil {
		slog.Error("Failed to get links", "error", err.Error(), "chatId", id)

		return errorText
	}

	if len(list) == 0 {
		responseText := "Список отслеживаемых ссылок пуст. Добавьте ссылки с помощью /track"
		return responseText
	}

	responseText := listLinksToString(list)

	return responseText
}

func (bot *Bot) stateWaitLink(id int64, text string) string {
	link := text
	if !validateLink(link) {
		bot.mu.Lock()
		delete(bot.userState, id)
		bot.mu.Unlock()

		responseText := "Поддерживается только gitHub(https://github.com/{owner}/{repo}) и " +
			"stackOverflow(https://stackoverflow.com/questions/{id}). Повторите команду /track"

		return responseText
	}

	bot.mu.Lock()
	bot.userState[id] = LinkWithState{
		Link:  domain.Link{URL: link},
		state: WaitingTags}
	bot.mu.Unlock()

	responseText := "Отправьте теги разделённые пробелами. Если не хотите добавлять теги введите \"-\" без кавычек "

	return responseText
}

func (bot *Bot) stateWaitTags(id int64, text string) string {
	bot.mu.RLock()
	curLinkWithState := bot.userState[id]
	bot.mu.RUnlock()

	if text == "-" {
		curLinkWithState.Link.Tags = []string{}
	} else {
		curLinkWithState.Link.Tags = strings.Split(text, " ")
	}

	curLinkWithState.state = WaitingFilters

	bot.mu.Lock()
	bot.userState[id] = curLinkWithState
	bot.mu.Unlock()

	responseText := "Отправьте фильтры разделённые пробелами. Если не хотите добавлять фильтры введите '-' без кавычек "

	return responseText
}

func (bot *Bot) stateWaitFilters(id int64, text string) string {
	bot.mu.RLock()
	curLinkWithState := bot.userState[id]
	bot.mu.RUnlock()

	if text == "-" {
		curLinkWithState.Link.Filters = []string{}
	} else {
		curLinkWithState.Link.Filters = strings.Split(text, " ")
	}

	curLinkWithState.state = NotState

	err := bot.scrapperHTTPClient.AddLink(id, curLinkWithState.Link)
	if err != nil {
		slog.Error(err.Error())

		if errors.As(err, &domain.ErrAPI{}) && (err.(domain.ErrAPI).ExceptionMessage == domain.ErrLinkAlreadyTracking{}.Error()) {
			responseText := "Данная ссылка уже отслеживается"
			return responseText
		}

		return errorText
	}

	bot.mu.Lock()
	delete(bot.userState, id)
	bot.mu.Unlock()

	responseText := "Ссылка отслеживается"

	return responseText
}

func (bot *Bot) stateWaitDelete(id int64, text string) string {
	link := text

	err := bot.scrapperHTTPClient.RemoveLink(id, domain.Link{URL: link})
	if err != nil {
		slog.Error(err.Error())

		return errorText
	}

	bot.mu.Lock()
	delete(bot.userState, id)
	bot.mu.Unlock()

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
