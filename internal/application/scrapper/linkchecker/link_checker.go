package linkchecker

import (
	"context"
	"log/slog"
	"net/url"
	"time"

	"LinkTracker/internal/application/scrapper"
	"LinkTracker/internal/domain"
)

type LinkSourceHandler interface {
	Supports(link *url.URL) bool
	Check(ctx context.Context, link *domain.Link) (lastUpdate time.Time, description string, err error)
}

type LinkChecker struct {
	linkRepo    scrapper.LinkRepo
	handlers    []LinkSourceHandler
	linkUpdates chan domain.LinkUpdate
}

func NewLinkChecker(linkRepo scrapper.LinkRepo, handlers []LinkSourceHandler) *LinkChecker {
	return &LinkChecker{
		linkRepo:    linkRepo,
		handlers:    handlers,
		linkUpdates: make(chan domain.LinkUpdate),
	}
}

func (l *LinkChecker) Scrape(ctx context.Context) {
	slog.Info("Scrape start")

	countChecks := 0
	successfullyChecks := 0

	links, err := l.linkRepo.GetAllLinks(ctx)
	if err != nil {
		slog.Error("links could not be retrieved", "error", err.Error())
		return
	}

	for _, link := range links {
		handler := l.findHandler(link.URL)
		if handler == nil {
			slog.Error("Unsupported host", "link", link.URL)
			continue
		}

		countChecks++

		lastUpdate, description, err := handler.Check(ctx, &link)
		if err != nil {
			slog.Error("Failed to check for updates on the link", "error", err.Error(), "link", link.URL)
			continue
		}

		successfullyChecks++

		if lastUpdate.After(link.LastUpdated) {
			err = l.linkRepo.UpdateTimeLink(ctx, lastUpdate, link.ID)
			if err != nil {
				slog.Error("Update time link failed", "error", err.Error(), "link", link.URL)
			}

			tgIDs, err := l.linkRepo.GetUsersByLink(ctx, link.ID)
			if err != nil {
				slog.Error("Failed to get users", "error", err.Error(), "link", link.URL)
				continue
			}

			l.linkUpdates <- domain.LinkUpdate{Link: link, TgIDs: tgIDs, Description: description}
		}
	}

	slog.Info("Scrape finished", "countChecks", countChecks, "successfullyChecks", successfullyChecks)
}

func (l *LinkChecker) findHandler(rawURL string) LinkSourceHandler {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}

	for _, h := range l.handlers {
		if h.Supports(parsed) {
			return h
		}
	}

	return nil
}
