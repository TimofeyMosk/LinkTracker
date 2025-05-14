package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"LinkTracker/internal/domain"

	"golang.org/x/time/rate"
)

const (
	githubAPIBaseURL       = "https://api.github.com"
	githubHTTPTimeout      = 5 * time.Second
	allowedRequestsPerHour = 60
)

// GitHubHTTPClient используется для работы с API GitHub.
type GitHubHTTPClient struct {
	Client        *http.Client
	globalLimiter *rate.Limiter
}

// NewGitHubHTTPClient создаёт нового клиента с заданным timeout.
func NewGitHubHTTPClient() *GitHubHTTPClient {
	return &GitHubHTTPClient{Client: &http.Client{
		Timeout: githubHTTPTimeout},
		globalLimiter: rate.NewLimiter(rate.Every(time.Hour/allowedRequestsPerHour), allowedRequestsPerHour)}
}

func (c *GitHubHTTPClient) Supports(link *url.URL) bool {
	return link.Host == "github.com"
}

func (c *GitHubHTTPClient) Check(ctx context.Context, link *domain.Link) (lastUpdate time.Time, description string, err error) {
	err = c.globalLimiter.Wait(ctx)
	if err != nil {
		slog.Error("Rate limit error", "error", err.Error())
		return time.Time{}, "", err
	}

	return c.GetLatestPROrIssue(ctx, link.URL)
}

// Ожидается формат: github.com/{owner}/{repo}.
func apiGitURLGeneration(link string) (string, error) {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("wrong url format, expected github.com/{owner}/{repo}")
	}

	owner, repo := parts[0], parts[1]
	apiURL := fmt.Sprintf("%s/repos/%s/%s", githubAPIBaseURL, owner, repo)

	return apiURL, nil
}

// GitHubIssue представляет Issue или Pull Request из GitHub API.
type GitHubIssue struct {
	Title string `json:"title"`
	User  struct {
		Login string `json:"login"`
	} `json:"user"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

// GetLatestPROrIssue возвращает сообщение с данными о последнем PR или Issue:
// название, имя пользователя, время создания и превью описания (200 символов).
// Пример ссылки: "https://github.com/TimofeyMosk/fractalFlame-image-creator"
func (c *GitHubHTTPClient) GetLatestPROrIssue(ctx context.Context, link string) (lastUpdate time.Time, description string, err error) {
	apiURL, err := apiGitURLGeneration(link)
	if err != nil {
		return time.Time{}, "", err
	}

	// Формируем URL для получения списка issues, сортируем по дате создания (от новых к старым), выбираем только один элемент.
	issuesURL := fmt.Sprintf("%s/issues?sort=created&direction=desc&per_page=1", apiURL)

	request, err := http.NewRequestWithContext(ctx, "GET", issuesURL, http.NoBody)
	if err != nil {
		return time.Time{}, "", err
	}

	response, err := c.Client.Do(request)
	if err != nil {
		return time.Time{}, "", err
	}

	defer func() {
		if cerr := response.Body.Close(); cerr != nil {
			slog.Error("could not close resource", "error", cerr.Error())
		}
	}()

	var issues []GitHubIssue
	if err := json.NewDecoder(response.Body).Decode(&issues); err != nil {
		return time.Time{}, "", err
	}

	if len(issues) == 0 {
		return time.Time{}, "", domain.ErrUpdatesNotFound{}
	}

	issue := issues[0]

	lastUpdate, err = time.Parse(time.RFC3339, issue.CreatedAt)
	if err != nil {
		return time.Time{}, "", err
	}

	description = createDescription(issue)

	return lastUpdate, description, nil
}

func createDescription(issue GitHubIssue) string {
	preview := issue.Body
	if len(preview) > 200 {
		preview = preview[:200]
	}

	description := fmt.Sprintf("Title: %s\nUser: %s\nCreated At: %s\nPreview: %s",
		issue.Title,
		issue.User.Login,
		issue.CreatedAt,
		preview,
	)

	return description
}
