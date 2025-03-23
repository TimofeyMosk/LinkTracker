package clients

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	githubAPIBaseURL  = "https://api.github.com"
	githubHTTPTimeout = 5 * time.Second
)

type GitHubHTTPClient struct {
	Client *http.Client
}

func NewGitHubHTTPClient() *GitHubHTTPClient {
	return &GitHubHTTPClient{Client: &http.Client{Timeout: githubHTTPTimeout}}
}

type GitHubRepoResponse struct {
	UpdatedAt string `json:"updated_at"`
}

// "https://github.com/TimofeyMosk/fractalFlame-image-creator"

func (c *GitHubHTTPClient) GetLastUpdateTimeRepo(link string) (time.Time, error) {
	apiURL, err := apiGitURLGeneration(link)
	if err != nil {
		return time.Time{}, err
	}

	request, err := http.NewRequest("GET", apiURL, http.NoBody)
	if err != nil {
		return time.Time{}, err
	}

	response, err := c.Client.Do(request)
	if err != nil {
		return time.Time{}, err
	}

	defer func() {
		if Cerr := response.Body.Close(); Cerr != nil {
			slog.Error("could not close resource", "error", Cerr.Error())
		}
	}()

	var repoData GitHubRepoResponse
	if err := json.NewDecoder(response.Body).Decode(&repoData); err != nil {
		return time.Time{}, err
	}

	updatedAt, err := time.Parse(time.RFC3339, repoData.UpdatedAt)
	if err != nil {
		return time.Time{}, err
	}

	return updatedAt, nil
}

func apiGitURLGeneration(link string) (string, error) {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("wrong url format, expected  github.com/{owner}/{repo}")
	}

	owner, repo := parts[0], parts[1]
	apiURL := fmt.Sprintf("%s/repos/%s/%s", githubAPIBaseURL, owner, repo)

	return apiURL, err
}
