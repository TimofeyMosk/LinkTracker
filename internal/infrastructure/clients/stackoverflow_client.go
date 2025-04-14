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
	stackOverflowAPIBaseURL  = "https://api.stackexchange.com/2.3"
	stackOverflowHTTPTimeout = 5 * time.Second
	allowedRequestsPerDay    = 3333
)

type StackOverflowHTTPClient struct {
	Client        *http.Client
	globalLimiter *rate.Limiter
}

func NewStackOverflowHTTPClient() *StackOverflowHTTPClient {
	return &StackOverflowHTTPClient{
		Client:        &http.Client{Timeout: stackOverflowHTTPTimeout},
		globalLimiter: rate.NewLimiter(rate.Every((24*time.Hour)/allowedRequestsPerDay), allowedRequestsPerDay)}
}

func (c *StackOverflowHTTPClient) Supports(link *url.URL) bool {
	return link.Host == "stackoverflow.com"
}

func (c *StackOverflowHTTPClient) Check(ctx context.Context, link *domain.Link) (lastUpdate time.Time, description string, err error) {
	err = c.globalLimiter.Wait(ctx)
	if err != nil {
		slog.Error("Rate limit error", "error", err.Error())
		return time.Time{}, "", err
	}

	return c.GetLatestAnswerOrComment(ctx, link.URL)
}

// SOQuestion представляет данные вопроса из StackOverflow API.
type SOQuestion struct {
	Title string `json:"title"`
}

// SOPost представляет общий тип для ответа или комментария.
type SOPost struct {
	Owner struct {
		DisplayName string `json:"display_name"`
	} `json:"owner"`
	CreationDate int64  `json:"creation_date"`
	Body         string `json:"body"`
}

// SOListResponse используется для декодирования списков ответов или комментариев.
type SOListResponse struct {
	Items []SOPost `json:"items"`
}

// extractQuestionID извлекает ID вопроса из ссылки.
// Ожидаемый формат: stackoverflow.com/questions/{id}/...
func extractQuestionID(link string) (string, error) {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != "questions" {
		return "", domain.ErrWrongURL{}
	}

	return parts[1], nil
}

// getQuestionDetails получает данные вопроса для получения заголовка.
func (c *StackOverflowHTTPClient) getQuestionDetails(ctx context.Context, questionID string) (SOQuestion, error) {
	apiURL := fmt.Sprintf("%s/questions/%s?site=stackoverflow", stackOverflowAPIBaseURL, questionID)

	request, err := http.NewRequestWithContext(ctx, "GET", apiURL, http.NoBody)
	if err != nil {
		return SOQuestion{}, err
	}

	response, err := c.Client.Do(request)
	if err != nil {
		return SOQuestion{}, err
	}

	defer func() {
		if cerr := response.Body.Close(); cerr != nil {
			slog.Error("could not close resource", "error", cerr.Error())
		}
	}()

	if response.StatusCode != http.StatusOK {
		return SOQuestion{}, domain.ErrStatusNotOK{StatusCode: response.StatusCode}
	}

	var result struct {
		Items []SOQuestion `json:"items"`
	}

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return SOQuestion{}, err
	}

	if len(result.Items) == 0 {
		return SOQuestion{}, domain.ErrWrongURL{}
	}

	return result.Items[0], nil
}

func (c *StackOverflowHTTPClient) getLatestByTag(ctx context.Context, questionID, tag string) (*SOPost, error) {
	apiURL := fmt.Sprintf("%s/questions/%s/%s?order=desc&sort=creation&site=stackoverflow&filter=withbody",
		stackOverflowAPIBaseURL, questionID, tag)

	request, err := http.NewRequestWithContext(ctx, "GET", apiURL, http.NoBody)
	if err != nil {
		return nil, err
	}

	response, err := c.Client.Do(request)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cerr := response.Body.Close(); cerr != nil {
			slog.Error("could not close resource", "error", cerr.Error())
		}
	}()

	if response.StatusCode != http.StatusOK {
		return nil, domain.ErrStatusNotOK{StatusCode: response.StatusCode}
	}

	var answerResp SOListResponse
	if err := json.NewDecoder(response.Body).Decode(&answerResp); err != nil {
		return nil, err
	}

	if len(answerResp.Items) == 0 {
		return nil, nil // ответов нет
	}

	return &answerResp.Items[0], nil
}

// GetLatestAnswerOrComment возвращает сообщение с данными о последнем ответе или комментарии к вопросу:
// заголовок вопроса, имя автора, время создания и превью текста (200 символов).
// Пример ссылки: "https://stackoverflow.com/questions/79467368/horizontal-scroll-component-does-not-work-as-expected-with-overflow"
func (c *StackOverflowHTTPClient) GetLatestAnswerOrComment(ctx context.Context, link string) (
	lastUpdate time.Time, description string, err error) {
	questionID, err := extractQuestionID(link)
	if err != nil {
		return time.Time{}, "", err
	}

	// Получаем заголовок вопроса.
	question, err := c.getQuestionDetails(ctx, questionID)
	if err != nil {
		return time.Time{}, "", err
	}

	// Пытаемся получить последний ответ.
	latestAnswer, err := c.getLatestByTag(ctx, questionID, "answers")
	if err != nil {
		return time.Time{}, "", err
	}

	// Если ответов нет, пробуем получить последний комментарий.
	var latestPost *SOPost
	if latestAnswer != nil {
		latestPost = latestAnswer
	} else {
		latestComment, err := c.getLatestByTag(ctx, questionID, "comments")
		if err != nil {
			return time.Time{}, "", err
		}

		if latestComment == nil {
			return time.Time{}, "", domain.ErrUpdatesNotFound{}
		}

		latestPost = latestComment
	}

	lastUpdate = time.Unix(latestPost.CreationDate, 0)
	description = createSODescription(question.Title, *latestPost)

	return lastUpdate, description, nil
}

// createSODescription формирует строку с информацией о последнем ответе или комментарии.
func createSODescription(questionTitle string, post SOPost) string {
	preview := post.Body
	if len(preview) > 200 {
		preview = preview[:200]
	}

	return fmt.Sprintf("Question: %s\nUser: %s\nCreated At: %s\nPreview: %s",
		questionTitle,
		post.Owner.DisplayName,
		time.Unix(post.CreationDate, 0).Format(time.RFC1123),
		preview,
	)
}
