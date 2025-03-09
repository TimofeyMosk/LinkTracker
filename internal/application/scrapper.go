package application

import (
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
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

func NewScrapper(db Database, interval time.Duration, botClient BotClient) *Scrapper {
	slog.Info("Creating new Scrapper", "interval", interval)

	return &Scrapper{
		db:         db,
		interval:   interval,
		stopSignal: make(chan struct{}),
		botClient:  botClient,
	}
}

func (s *Scrapper) Run() error {
	sched, err := gocron.NewScheduler()
	if err != nil {
		slog.Error("Failed to create scheduler", "error", err.Error())
		return fmt.Errorf("could not create sheduler: %w", err)
	}

	_, err = sched.NewJob(
		gocron.DurationJob(s.interval),
		gocron.NewTask(s.Scrape),
	)

	if err != nil {
		slog.Error("Failed to create job", "error", err.Error())
		return fmt.Errorf("could not create job: %w", err)
	}

	slog.Info("Starts scrapper")
	sched.Start()

	<-s.stopSignal
	slog.Info("Shutting down scrapper")

	err = sched.Shutdown()
	if err != nil {
		slog.Error("Failed to shutdown scrapper", "error", err.Error())
		return fmt.Errorf("could not shutdown scrapper: %w", err)
	}

	return nil
}

func (s *Scrapper) Stop() error {
	s.stopSignal <- struct{}{}
	return nil
}

func (s *Scrapper) Scrape() {
	slog.Info("Scrape start")

	countChecks := 0
	successfullyChecks := 0

	usersID, err := s.db.GetAllUsers()
	if err != nil {
		slog.Error("Failed to retrieve the list of users", "error", err.Error())
		return
	}

	for _, tgID := range usersID {
		links, err := s.db.GetLinks(tgID)
		if err != nil {
			slog.Error("User links could not be retrieved", "error", err.Error(), "tgID", tgID)
		}

		for _, link := range links {
			countChecks++

			activity, err := CheckUpdates(link.URL, time.Now().Add(-5*time.Minute))
			if err != nil {
				slog.Error("Failed to check for updates on the link", "error", err.Error(), "link", link.URL)
				continue
			}

			successfullyChecks++

			if activity {
				err := s.botClient.PostUpdates(link, tgID)
				if err != nil {
					slog.Error("Failed to send user updates", "error", err.Error(), "link", link.URL)
				}
			}
		}
	}

	slog.Info("Scrape finished", "countChecks", countChecks, "successfullyChecks", successfullyChecks)
}

func (s *Scrapper) AddUser(id int64) error {
	slog.Info("Adding user")

	err := s.db.CreateUser(id)
	if err != nil {
		slog.Error("Failed to add user", "error", err.Error(), "tgID", id)
		return err
	}

	return nil
}

func (s *Scrapper) DeleteUser(id int64) error {
	slog.Info("Deleting user", "tgID", id)

	err := s.db.DeleteUser(id)
	if err != nil {
		slog.Error("Failed to delete user", "error", err.Error(), "tgID", id)
		return err
	}

	return nil
}

func (s *Scrapper) GetLinks(id int64) ([]domain.Link, error) {
	slog.Info("Getting links", "tgID", id)

	links, err := s.db.GetLinks(id)
	if err != nil {
		slog.Error("Failed to get links", "error", err.Error(), "tgID", id)
		return nil, err
	}

	return links, nil
}

func (s *Scrapper) AddLink(id int64, link domain.Link) (domain.Link, error) {
	slog.Info("Adding link", "tgID", id, "link", link)

	if !validLink(link.URL) {
		slog.Error("Invalid link URL", "tgID", id, "link", link)
		return link, domain.ErrWrongURL{}
	}

	err := s.db.AddLink(id, link)
	if err != nil {
		slog.Error("Failed to add link", "error", err.Error(), "tgID", id, "link", link)
		return link, err
	}

	return link, nil
}

func (s *Scrapper) DeleteLink(id int64, link domain.Link) (domain.Link, error) {
	slog.Info("Deleting link", "tgID", id, "link", link)

	deletedLink, err := s.db.DeleteLink(id, link)
	if err != nil {
		slog.Error("Failed to remove link", "error", err.Error(), "tgID", id, "link", link)
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
