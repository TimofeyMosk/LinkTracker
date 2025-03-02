package application

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/requests"
)

type SOResponse struct {
	Items []struct {
		LastActivityDate int64 `json:"last_activity_date"`
	} `json:"items"`
}

// https://stackoverflow.com/questions/79467368/horizontal-scroll-component-does-not-work-as-expected-with-overflow

func CheckStackOverflowQuestionUpdate(link string, lastKnown time.Time) (bool, time.Time, error) {
	apiURL, err := apiSOUrlGeneration(link)
	if err != nil {
		return false, lastKnown, err
	}

	resp, err := requests.GetRequest(apiURL)
	if err != nil {
		return false, lastKnown, err
	}

	defer func() {
		Cerr := resp.Body.Close()
		if Cerr != nil {
			slog.Error("failed to close response: %v\n", "error", Cerr)
		}
	}()

	var soResponse SOResponse
	if err := json.NewDecoder(resp.Body).Decode(&soResponse); err != nil {
		return false, lastKnown, err
	}

	if len(soResponse.Items) == 0 {
		return false, lastKnown, ErrQuestionNotFound{}
	}

	lastActivity := time.Unix(soResponse.Items[0].LastActivityDate, 0)

	if lastActivity.After(lastKnown) {
		return true, lastActivity, nil
	}

	return false, lastKnown, nil
}

func apiSOUrlGeneration(link string) (apiURL string, err error) {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != "questions" {
		return "", fmt.Errorf("wrong url format, expected   stackoverflow.com/questions/{id}")
	}

	questionID := parts[1]
	apiURL = fmt.Sprintf("https://api.stackexchange.com/2.2/questions/%s?order=desc&sort=activity&site=stackoverflow", questionID)

	return
}
