package application

import (
	"log/slog"
	"net/url"
	"time"

	"LinkTracker/internal/domain"
	"LinkTracker/internal/infrastructure/clients"
)

func CheckUpdates(linkURL string, lastKnown time.Time) (bool, error) {
	parsedURL, err := url.Parse(linkURL)
	if err != nil {
		return false, err
	}

	const (
		github        = "github.com"
		stackoverflow = "stackoverflow.com"
	)

	switch parsedURL.Host {
	case github:
		gitClient := clients.NewGitHubHTTPClient()

		lastUpdate, err := gitClient.GetLastUpdateTimeRepo(linkURL)
		if err != nil {
			slog.Error(err.Error(), "linkURL", linkURL)
			return false, err
		}

		return lastUpdate.After(lastKnown), nil
	case stackoverflow:
		soClient := clients.NewStackOverflowHTTPClient()

		lastActivity, err := soClient.GetLastActivityQuestion(linkURL)
		if err != nil {
			slog.Error(err.Error(), "linkURL", linkURL)
			return false, err
		}

		return lastActivity.After(lastKnown), nil
	default:
		slog.Error("Unsupported host", "host", parsedURL.Host)
		return false, domain.ErrUnsupportedHost{}
	}
}
