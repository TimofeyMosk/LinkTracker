package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"LinkTracker/internal/domain"
	"LinkTracker/pkg"
)

const (
	stackOverflowAPIBaseURL  = "https://api.stackexchange.com/2.2"
	stackOverflowHTTPTimeout = 5 * time.Second
)

type StackOverflowHTTPClient struct {
	Client *http.Client
}

func NewStackOverflowHTTPClient() *StackOverflowHTTPClient {
	return &StackOverflowHTTPClient{
		Client: &http.Client{Timeout: stackOverflowHTTPTimeout},
	}
}

type SOResponse struct {
	Items []struct {
		LastActivityDate int64 `json:"last_activity_date"`
	} `json:"items"`
}

//https://stackoverflow.com/questions/79467368/horizontal-scroll-component-does-not-work-as-expected-with-overflow

func (c *StackOverflowHTTPClient) GetLastActivityQuestion(link string) (time.Time, error) {
	apiURL, err := apiSOUrlGeneration(link)
	if err != nil {
		return time.Time{}, err
	}

	request, err := http.NewRequest("GET", apiURL, http.NoBody)
	if err != nil {
		return time.Time{}, err
	}

	response, err := c.Client.Do(request) //nolint:bodyclose // The body closes in a function pkg.SafeClose(response.Body)
	if err != nil {
		return time.Time{}, err
	}
	defer pkg.SafeClose(response.Body)

	var soResponse SOResponse
	if err := json.NewDecoder(response.Body).Decode(&soResponse); err != nil {
		return time.Time{}, err
	}

	if len(soResponse.Items) == 0 {
		return time.Time{}, domain.ErrQuestionNotFound{}
	}

	lastActivity := time.Unix(soResponse.Items[0].LastActivityDate, 0)

	return lastActivity, nil
}

func apiSOUrlGeneration(link string) (string, error) {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != "questions" {
		return "", fmt.Errorf("wrong url format, expected   stackoverflow.com/questions/{id}")
	}

	questionID := parts[1]
	apiURL := fmt.Sprintf("%s/questions/%s?order=desc&sort=activity&site=stackoverflow", stackOverflowAPIBaseURL, questionID)

	return apiURL, err
}
