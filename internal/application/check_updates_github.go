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

type GitHubRepoResponse struct {
	UpdatedAt string `json:"updated_at"`
}

// "https://github.com/TimofeyMosk/fractalFlame-image-creator"

func CheckGitHubRepoUpdate(link string, lastKnown time.Time) (bool, time.Time, error) {
	apiURL, err := apiGitURLGeneration(link)
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

	var repoData GitHubRepoResponse
	if err := json.NewDecoder(resp.Body).Decode(&repoData); err != nil {
		return false, lastKnown, err
	}

	updatedAt, err := time.Parse(time.RFC3339, repoData.UpdatedAt)
	if err != nil {
		return false, lastKnown, err
	}

	if !updatedAt.After(lastKnown) {
		return false, lastKnown, nil
	}

	return true, updatedAt, nil
}

func apiGitURLGeneration(link string) (apiURL string, err error) {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("wrong url format, expected  github.com/{owner}/{repo}")
	}

	owner, repo := parts[0], parts[1]
	apiURL = fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)

	return
}
