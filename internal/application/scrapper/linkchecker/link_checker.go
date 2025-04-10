package linkchecker

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"LinkTracker/internal/application/scrapper"
	"LinkTracker/internal/domain"
)

// LinkSourceHandler определяет интерфейс для проверки ссылки для конкретного источника.
type LinkSourceHandler interface {
	Supports(link *url.URL) bool
	Check(ctx context.Context, link *domain.Link) (lastUpdate time.Time, description string, err error)
}

// LinkChecker выполняет проверку ссылок в пакетном и параллельном режимах.
type LinkChecker struct {
	linkRepo         scrapper.LinkRepo
	handlers         []LinkSourceHandler
	limitLinksInPage int64
	workers          int
}

// NewLinkChecker создаёт новый экземпляр LinkChecker.
func NewLinkChecker(linkRepo scrapper.LinkRepo, handlers []LinkSourceHandler, limitLinksInPage int64, workers int) *LinkChecker {
	if workers < 1 {
		workers = 1
	}

	return &LinkChecker{
		linkRepo:         linkRepo,
		handlers:         handlers,
		limitLinksInPage: limitLinksInPage,
		workers:          workers,
	}
}

// CheckLinks выполняет обход ссылок пакетами с параллельной обработкой каждого батча.
// Обновления передаются через канал linkUpdates.
func (l *LinkChecker) CheckLinks(ctx context.Context, linkUpdates chan<- domain.LinkUpdate) {
	slog.Info("Scrape start")

	var (
		lastUpdateTime   time.Time
		totalChecks      int64
		successfulChecks int64
	)

	// Цикл по батчам ссылок через курсорную пагинацию.
	for {
		links, err := l.linkRepo.GetLinksAfter(ctx, lastUpdateTime, l.limitLinksInPage)
		if err != nil {
			slog.Error("Failed to retrieve links", "error", err.Error())
			return
		}

		if len(links) == 0 {
			break
		}

		// Обновляем курсор – берем время обновления последней ссылки из батча
		lastUpdateTime = links[len(links)-1].LastUpdated

		chunks := partitionLinks(links, l.workers)

		var wg sync.WaitGroup

		wg.Add(len(chunks))

		for _, chunk := range chunks {
			go func(chunk []domain.Link) {
				defer wg.Done()

				for _, link := range chunk {
					atomic.AddInt64(&totalChecks, 1)

					err := l.processLink(ctx, &link, linkUpdates, &successfulChecks)
					if err != nil {
						slog.Error("Error processing link", "link", link.URL, "error", err.Error())
					}
				}
			}(chunk)
		}

		wg.Wait()
	}

	slog.Info("Scrape finished",
		"totalChecks", totalChecks,
		"successfulChecks", successfulChecks,
	)
}

// processLink обрабатывает одну ссылку: ищет подходящий обработчик,
// запускает проверку и, если необходимо, обновляет время последнего обновления и отправляет обновление через канал.
func (l *LinkChecker) processLink(ctx context.Context, link *domain.Link,
	linkUpdates chan<- domain.LinkUpdate, successfulChecks *int64) error {
	handler := l.findHandler(link.URL)
	if handler == nil {
		slog.Error("Unsupported host", "link", link.URL)
		return domain.ErrUnsupportedHost{}
	}

	lastUpdate, description, err := handler.Check(ctx, link)
	if err != nil {
		return err
	}

	err = l.linkRepo.UpdateTimeLink(ctx, lastUpdate, link.ID)
	if err != nil {
		slog.Error("Update time link failed", "error", err.Error(), "link", link.URL)
		return fmt.Errorf("failed update time: %w", err)
	}

	atomic.AddInt64(successfulChecks, 1)

	if lastUpdate.After(link.LastUpdated) {
		tgIDs, err := l.linkRepo.GetUsersByLink(ctx, link.ID)
		if err != nil {
			slog.Error("Failed to get users", "error", err.Error(), "link", link.URL)
			return fmt.Errorf("failed to get users: %w", err)
		}

		linkUpdates <- domain.LinkUpdate{
			Link:        *link,
			TgIDs:       tgIDs,
			Description: description,
		}
	}

	return nil
}

// findHandler парсит URL и ищет первый подходящий обработчик.
func (l *LinkChecker) findHandler(rawURL string) LinkSourceHandler {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		slog.Error("Failed to parse URL", "url", rawURL, "error", err.Error())
		return nil
	}

	for _, h := range l.handlers {
		if h.Supports(parsed) {
			return h
		}
	}

	return nil
}

// partitionLinks делит срез links на n равномерных частей.
func partitionLinks(links []domain.Link, n int) [][]domain.Link {
	chunkSize := len(links)/n + 1
	chunks := make([][]domain.Link, 0, n)

	for i := 0; i < len(links); i += chunkSize {
		end := i + chunkSize
		if end > len(links) {
			end = len(links)
		}

		chunks = append(chunks, links[i:end])
	}

	return chunks
}
