package scrapper

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"LinkTracker/internal/domain"

	"github.com/go-co-op/gocron/v2"
)

type LinkRepo interface {
	GetUserLinks(ctx context.Context, tgID int64) ([]domain.Link, error)
	AddLink(ctx context.Context, tgID int64, link *domain.Link) (domain.Link, error)
	DeleteLink(ctx context.Context, tgID int64, link *domain.Link) (domain.Link, error)
	UpdateLink(ctx context.Context, tgID int64, link *domain.Link) error
	GetAllLinks(ctx context.Context) ([]domain.Link, error)
	GetUsersByLink(ctx context.Context, linkID int64) ([]int64, error)
	UpdateTimeLink(ctx context.Context, lastUpdate time.Time, linkID int64) error
	GetLinksAfter(ctx context.Context, lastUpdate time.Time, limit int64) ([]domain.Link, error)
}

type UserRepo interface {
	CreateUser(ctx context.Context, id int64) error
	DeleteUser(ctx context.Context, id int64) error
	GetAllUsers(ctx context.Context) ([]int64, error)
}

type StateRepo interface {
	CreateState(ctx context.Context, tgID int64, state int) error
	DeleteState(ctx context.Context, tgID int64) error
	GetState(ctx context.Context, tgID int64) (state int, link domain.Link, err error)
	UpdateState(ctx context.Context, tgID int64, state int, link *domain.Link) error
}

type Notifier interface {
	PostUpdates(ctx context.Context, link *domain.Link, tgID []int64, description string) error
}

type LinkChecker interface {
	CheckLinks(ctx context.Context, linkUpdates chan<- domain.LinkUpdate)
}

type Scrapper struct {
	userRepo     UserRepo
	linkRepo     LinkRepo
	stateManager StateRepo
	notifier     Notifier
	linkCheck    LinkChecker
	interval     time.Duration
	linkUpdates  chan domain.LinkUpdate
}

func NewScrapper(userRepo UserRepo, linkRepo LinkRepo, stateManager StateRepo,
	interval time.Duration, notifier Notifier, linkChecker LinkChecker) *Scrapper {
	linkUpdatesBufferSize := 1000

	slog.Info("Creating new Scrapper", "interval", interval, "linkUpdatesBufferSize", linkUpdatesBufferSize)

	return &Scrapper{
		userRepo:     userRepo,
		linkRepo:     linkRepo,
		stateManager: stateManager,
		interval:     interval,
		notifier:     notifier,
		linkCheck:    linkChecker,
		linkUpdates:  make(chan domain.LinkUpdate, linkUpdatesBufferSize),
	}
}

func (s *Scrapper) Run(ctx context.Context) error {
	scheduler, err := initLinksCheckerScheduler(ctx, s.interval, s.linkCheck.CheckLinks, s.linkUpdates)
	if err != nil {
		return err
	}

	slog.Info("Starts scrapper scheduler")
	scheduler.Start()

	go func() {
		for update := range s.linkUpdates {
			err := s.notifier.PostUpdates(ctx, &update.Link, update.TgIDs, update.Description)
			if err != nil {
				slog.Error(err.Error(), "url", update.Link, "tgIDS", update.TgIDs, "description", update.Description)
			}
		}
	}()

	<-ctx.Done()
	close(s.linkUpdates)
	slog.Info("Shutting down scrapper")

	err = scheduler.Shutdown()
	if err != nil {
		slog.Error("Failed to shutdown scrapper", "error", err.Error())
		return fmt.Errorf("could not shutdown scrapper: %w", err)
	}

	return nil
}

func (s *Scrapper) AddUser(ctx context.Context, tgID int64) error {
	err := s.userRepo.CreateUser(ctx, tgID)
	if err != nil {
		slog.Error("Add user failed", "error", err.Error(), "tgID", tgID)
		return err
	}

	slog.Info("Add user done", "tgID", tgID)

	return nil
}

func (s *Scrapper) DeleteUser(ctx context.Context, tgID int64) error {
	err := s.userRepo.DeleteUser(ctx, tgID)
	if err != nil {
		slog.Error("Delete user failed", "error", err.Error(), "tgID", tgID)
		return err
	}

	slog.Info("Delete user done", "tgID", tgID)

	return nil
}

func (s *Scrapper) GetUserLinks(ctx context.Context, tgID int64) ([]domain.Link, error) {
	links, err := s.linkRepo.GetUserLinks(ctx, tgID)
	if err != nil {
		slog.Error("Get user links failed", "error", err.Error(), "tgID", tgID)
		return nil, err
	}

	slog.Info("Get user links done", "tgID", tgID)

	return links, nil
}

func (s *Scrapper) AddLink(ctx context.Context, tgID int64, newLink *domain.Link) (domain.Link, error) {
	userLinks, err := s.linkRepo.GetUserLinks(ctx, tgID)
	if err != nil {
		slog.Error("Add link failed", "error", err.Error(), "tgID", tgID, "link", newLink.URL)
		return domain.Link{}, err
	}

	for _, userLink := range userLinks {
		if userLink.URL == newLink.URL {
			return domain.Link{}, domain.ErrLinkAlreadyTracking{}
		}
	}

	newLinkWithID, err := s.linkRepo.AddLink(ctx, tgID, newLink)
	if err != nil {
		slog.Error("Add link failed", "error", err.Error(), "tgID", tgID, "link", newLink.URL)
		return domain.Link{}, err
	}

	slog.Info("Add link done", "tgID", tgID, "link", newLink.URL)

	return newLinkWithID, nil
}

func (s *Scrapper) DeleteLink(ctx context.Context, tgID int64, link *domain.Link) (domain.Link, error) {
	deletedLink, err := s.linkRepo.DeleteLink(ctx, tgID, link)
	if err != nil {
		slog.Error("Delete link failed", "error", err.Error(), "tgID", tgID, "link", link.URL)

		if errors.Is(err, domain.ErrLinkNotExist{}) {
			return domain.Link{}, domain.ErrLinkNotExist{}
		}

		return domain.Link{}, err
	}

	slog.Info("Deleting link done", "tgID", tgID, "link", link.URL)

	return deletedLink, nil
}

func (s *Scrapper) UpdateLink(ctx context.Context, tgID int64, link *domain.Link) error {
	err := s.linkRepo.UpdateLink(ctx, tgID, link)
	if err != nil {
		slog.Error("Update link failed", "error", err.Error(), "tgID", tgID, "link", link.URL)
	}

	slog.Info("Update link done", "tgID", tgID, "link", link.URL)

	return err
}

func (s *Scrapper) CreateState(ctx context.Context, tgID int64, state int) error {
	err := s.stateManager.CreateState(ctx, tgID, state)
	if err != nil {
		slog.Error("Create state failed", "error", err.Error(), "tgID", tgID, "state", state)
	}

	slog.Info("Create state done", "tgID", tgID, "state", state)

	return err
}

func (s *Scrapper) DeleteState(ctx context.Context, tgID int64) error {
	err := s.stateManager.DeleteState(ctx, tgID)
	if err != nil {
		slog.Error("Delete state failed", "error", err.Error(), "tgID", tgID)
	}

	slog.Info("Deleting state done", "tgID", tgID)

	return err
}

func (s *Scrapper) GetState(ctx context.Context, tgID int64) (int, domain.Link, error) {
	state, link, err := s.stateManager.GetState(ctx, tgID)
	if err != nil {
		slog.Error("Get state failed", "error", err.Error(), "tgID", tgID)
		return -1, domain.Link{}, err
	}

	slog.Info("Get state done", "tgID", tgID, "state", state, "link", link.URL)

	return state, link, nil
}

func (s *Scrapper) UpdateState(ctx context.Context, tgID int64, state int, link *domain.Link) error {
	err := s.stateManager.UpdateState(ctx, tgID, state, link)
	if err != nil {
		slog.Error("Update state failed", "error", err.Error(), "tgID", tgID, "state", state, "link", link.URL)
	}

	slog.Info("Updating state done", "tgID", tgID, "state", state, "link", link.URL)

	return err
}

func initLinksCheckerScheduler(ctx context.Context, interval time.Duration,
	scrapeFunc func(ctx context.Context, updates chan<- domain.LinkUpdate), updates chan<- domain.LinkUpdate) (gocron.Scheduler, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		slog.Error("Failed to create scheduler", "error", err.Error())
		return nil, fmt.Errorf("could not create sheduler: %w", err)
	}

	_, err = scheduler.NewJob(
		gocron.DurationJob(interval),
		gocron.NewTask(func() {
			ctxWithTimeout, cancelTimeout := context.WithTimeout(ctx, interval)
			defer cancelTimeout()
			scrapeFunc(ctxWithTimeout, updates)
		}),
	)

	if err != nil {
		slog.Error("Failed to create job", "error", err.Error())
		return nil, fmt.Errorf("could not create job: %w", err)
	}

	return scheduler, nil
}
