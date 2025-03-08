package application

import (
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients"

	"github.com/go-co-op/gocron/v2"
)

type Database interface {
	CreateUser(id int64) error
	DeleteUser(id int64) error
	GetLinks(id int64) ([]domain.Link, error)
	AddLink(id int64, link domain.Link) error
	DeleteLink(id int64, link domain.Link) (domain.Link, error)
	GetAllLinks() ([]domain.Link, error)
	GetAllUsers() ([]int64, error)
}

type BotClient interface {
	PostUpdates(link domain.Link, tgID int64) error
}

type Scrapper struct {
	db         Database
	botClient  BotClient
	interval   time.Duration
	stopSignal chan struct{}
}

func NewScrapper(db Database, interval time.Duration, httpClient BotClient) *Scrapper {
	slog.Info("Creating new Scrapper", "interval", interval)

	return &Scrapper{
		db:         db,
		interval:   interval,
		stopSignal: make(chan struct{}),
		botClient:  httpClient,
	}
}

func (s *Scrapper) Run() error {
	sched, err := gocron.NewScheduler()
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("could not create sheduler: %w", err)
	}

	_, err = sched.NewJob(
		gocron.DurationJob(s.interval),
		gocron.NewTask(s.Scrape),
	)

	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("could not create job: %w", err)
	}

	slog.Info("Starts scrapper")
	sched.Start()

	<-s.stopSignal
	slog.Info("Shutting down scrapper")

	err = sched.Shutdown()
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("could not shutdown scrapper: %w", err)
	}

	return nil
}

func (s *Scrapper) Stop() error {
	s.stopSignal <- struct{}{}
	return nil
}

func (s *Scrapper) Scrape() {
	slog.Info("Scrape")

	usersID, err := s.db.GetAllUsers()
	if err != nil {
		slog.Error("Failed to retrieve the list of users", "error", err.Error())
		return
	}

	for _, tgID := range usersID {
		links, err := s.db.GetLinks(tgID)
		if err != nil {
			slog.Error("User links could not be retrieved", "error", err.Error())
		}

		for _, link := range links {
			activity, err := s.CheckUpdates(link.URL, time.Now().Add(-5*time.Minute))
			if err != nil {
				slog.Error(err.Error())
			}

			if activity {
				err := s.botClient.PostUpdates(link, tgID)
				if err != nil {
					slog.Error(err.Error(), "link", link.URL)
				}
			}
		}
	}
}

func (s *Scrapper) CheckUpdates(linkURL string, lastKnown time.Time) (bool, error) {
	parsedURL, err := url.Parse(linkURL)
	if err != nil {
		return false, err
	}

	switch parsedURL.Host {
	case "github.com":
		gitClient := clients.NewGitHubHTTPClient()

		lastUpdate, err := gitClient.GetLastUpdateTimeRepo(linkURL)
		if err != nil {
			slog.Error(err.Error(), "linkURL", linkURL)
			return false, err
		}

		return lastUpdate.After(lastKnown), nil
	case "stackoverflow.com":
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

func (s *Scrapper) AddUser(id int64) error {
	slog.Info("Adding user")

	err := s.db.CreateUser(id)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
}

func (s *Scrapper) DeleteUser(id int64) error {
	slog.Info("Deleting user")

	err := s.db.DeleteUser(id)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
}

func (s *Scrapper) GetLinks(id int64) ([]domain.Link, error) {
	slog.Info("Getting links")

	links, err := s.db.GetLinks(id)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return links, nil
}

func (s *Scrapper) AddLink(id int64, link domain.Link) (domain.Link, error) {
	slog.Info("Adding link", "link", link)

	if !validLink(link.URL) {
		slog.Error("Invalid link URL", "link", link)
		return link, domain.ErrWrongURL{}
	}

	err := s.db.AddLink(id, link)
	if err != nil {
		slog.Error(err.Error())
		return link, err
	}

	return link, nil
}

func (s *Scrapper) DeleteLink(id int64, link domain.Link) (domain.Link, error) {
	slog.Info("Deleting link", "link", link)

	deletedLink, err := s.db.DeleteLink(id, link)
	if err != nil {
		slog.Error(err.Error())
		return domain.Link{}, err
	}

	return deletedLink, nil
}

func validLink(linkURL string) bool {
	parsedURL, err := url.Parse(linkURL)
	if err != nil {
		return false
	}

	switch parsedURL.Host {
	case "github.com":
		return true
	case "stackoverflow.com":
		return true
	default:
		return false
	}
}
