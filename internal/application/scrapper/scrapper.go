package scrapper

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"LinkTracker/internal/domain"

	"github.com/go-co-op/gocron/v2"
)

type LinkRepo interface {
	GetUserLinks(ctx context.Context, tgID int64) ([]domain.Link, error)
	AddLink(ctx context.Context, tgID int64, link *domain.Link) error
	DeleteLink(ctx context.Context, tgID int64, link *domain.Link) (domain.Link, error)
	GetAllLinks(ctx context.Context) ([]domain.Link, error)
	GetUsersByLink(ctx context.Context, linkID int64) ([]int64, error)
	UpdateTimeLink(ctx context.Context, lastUpdate time.Time, linkID int64) error
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
	Scrape(ctx context.Context)
}

type Scrapper struct {
	userRepo     UserRepo
	linkRepo     LinkRepo
	stateManager StateRepo
	notifier     Notifier
	linkCheck    LinkChecker
	interval     time.Duration
}

func NewScrapper(userRepo UserRepo, linkRepo LinkRepo, stateManager StateRepo,
	interval time.Duration, notifier Notifier, linkChecker LinkChecker) *Scrapper {
	slog.Info("Creating new Scrapper", "interval", interval)

	return &Scrapper{
		userRepo:     userRepo,
		linkRepo:     linkRepo,
		stateManager: stateManager,
		interval:     interval,
		notifier:     notifier,
		linkCheck:    linkChecker,
	}
}

func (s *Scrapper) Run(ctx context.Context) error {
	scheduler, err := initScrapperScheduler(ctx, s.interval, s.linkCheck.Scrape)
	if err != nil {
		return err
	}

	slog.Info("Starts scrapper scheduler")
	scheduler.Start()

	<-ctx.Done()
	slog.Info("Shutting down scrapper")

	err = scheduler.Shutdown()
	if err != nil {
		slog.Error("Failed to shutdown scrapper", "error", err.Error())
		return fmt.Errorf("could not shutdown scrapper: %w", err)
	}

	return nil
}

func (s *Scrapper) AddUser(ctx context.Context, tgID int64) error {
	slog.Info("Adding user")

	err := s.userRepo.CreateUser(ctx, tgID)
	if err != nil {
		slog.Error("Failed to add user", "error", err.Error(), "tgID", tgID)
		return err
	}

	return nil
}

func (s *Scrapper) DeleteUser(ctx context.Context, tgID int64) error {
	slog.Info("Deleting user", "tgID", tgID)

	err := s.userRepo.DeleteUser(ctx, tgID)
	if err != nil {
		slog.Error("Failed to delete user", "error", err.Error(), "tgID", tgID)
		return err
	}

	return nil
}

func (s *Scrapper) GetUserLinks(ctx context.Context, tgID int64) ([]domain.Link, error) {
	slog.Info("Getting links", "tgID", tgID)

	links, err := s.linkRepo.GetUserLinks(ctx, tgID)
	if err != nil {
		slog.Error("Failed to get links", "error", err.Error(), "tgID", tgID)
		return nil, err
	}

	return links, nil
}

func (s *Scrapper) AddLink(ctx context.Context, tgID int64, newLink *domain.Link) (domain.Link, error) {
	slog.Info("Adding link", "tgID", tgID, "link", newLink)

	userLinks, err := s.linkRepo.GetUserLinks(ctx, tgID)
	if err != nil {
		slog.Error("Failed to get links", "error", err.Error(), "tgID", tgID)
		return *newLink, err
	}

	for _, userLink := range userLinks {
		if userLink.URL == newLink.URL {
			return *newLink, domain.ErrLinkAlreadyTracking{}
		}
	}

	err = s.linkRepo.AddLink(ctx, tgID, newLink)
	if err != nil {
		slog.Error("Failed to add link", "error", err.Error(), "tgID", tgID, "link", newLink)
		return *newLink, err
	}

	return *newLink, nil
}

func (s *Scrapper) DeleteLink(ctx context.Context, tgID int64, link *domain.Link) (domain.Link, error) {
	slog.Info("Deleting link", "tgID", tgID, "link", link)

	deletedLink, err := s.linkRepo.DeleteLink(ctx, tgID, link)
	if err != nil {
		slog.Error("Failed to remove link", "error", err.Error(), "tgID", tgID, "link", link)
		return domain.Link{}, err
	}

	return deletedLink, nil
}

func (s *Scrapper) CreateState(ctx context.Context, tgID int64, state int) error {
	slog.Info("Creating state")

	err := s.stateManager.CreateState(ctx, tgID, state)
	if err != nil {
		return err
	}

	return nil
}

func (s *Scrapper) DeleteState(ctx context.Context, tgID int64) error {
	slog.Info("Deleting state")

	err := s.stateManager.DeleteState(ctx, tgID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Scrapper) GetState(ctx context.Context, tgID int64) (int, domain.Link, error) {
	slog.Info("Getting state")

	state, link, err := s.stateManager.GetState(ctx, tgID)
	if err != nil {
		return -1, domain.Link{}, err
	}

	return state, link, nil
}

func (s *Scrapper) UpdateState(ctx context.Context, tgID int64, state int, link *domain.Link) error {
	slog.Info("Updating state")

	err := s.stateManager.UpdateState(ctx, tgID, state, link)
	if err != nil {
		return err
	}

	return nil
}

func initScrapperScheduler(ctx context.Context, interval time.Duration, scrapeFunc func(ctx context.Context)) (gocron.Scheduler, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		slog.Error("Failed to create scheduler", "error", err.Error())
		return nil, fmt.Errorf("could not create sheduler: %w", err)
	}

	_, err = scheduler.NewJob(
		gocron.DurationJob(interval),
		gocron.NewTask(func() {
			ctx, cancel := context.WithTimeout(ctx, interval)
			defer cancel()
			scrapeFunc(ctx)
		}),
		gocron.WithContext(ctx),
	)

	if err != nil {
		slog.Error("Failed to create job", "error", err.Error())
		return nil, fmt.Errorf("could not create job: %w", err)
	}

	return scheduler, nil
}
