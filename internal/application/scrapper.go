package application

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	botdto "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/dto_bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/requests"

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

type Scrapper struct {
	db         Database
	interval   time.Duration
	stopSignal chan struct{}
	botBaseURL string
}

func NewScrapper(db Database, interval time.Duration, botBaseURL string) *Scrapper {
	slog.Info("Creating new Scrapper", "interval", interval, "botBaseURL", botBaseURL)

	return &Scrapper{
		db:         db,
		interval:   interval,
		stopSignal: make(chan struct{}),
		botBaseURL: botBaseURL,
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
				resp, err := s.PostUpdates(link, tgID)
				if err != nil {
					slog.Error(err.Error())
				}

				if !(resp.StatusCode == http.StatusOK) {
					var errorResponse botdto.ApiErrorResponse
					_ = json.NewDecoder(resp.Body).Decode(&errorResponse)
					slog.Error("failed to post updates", "response", errorResponse)
				}

				err = resp.Body.Close()
				if err != nil {
					slog.Error("failed to close response: %v\n", "error", err)
				}
			}
		}
	}
}

func (s *Scrapper) PostUpdates(link domain.Link, tgID int64) (*http.Response, error) {
	linkUpdate := botdto.LinkUpdate{
		Description: nil,
		Id:          nil,
		TgChatIds:   &[]int64{tgID},
		Url:         &link.URL,
	}

	var data bytes.Buffer

	err := json.NewEncoder(&data).Encode(linkUpdate)
	if err != nil {
		return nil, err
	}

	return requests.PostRequest(s.botBaseURL+"/updates", &data)
}

func (s *Scrapper) CheckUpdates(linkURL string, lastKnown time.Time) (bool, error) {
	parsedURL, err := url.Parse(linkURL)
	if err != nil {
		return false, err
	}

	switch parsedURL.Host {
	case "github.com":
		isEvent, _, err := CheckGitHubRepoUpdate(linkURL, lastKnown)
		if err != nil {
			return false, err
		}

		return isEvent, nil
	case "stackoverflow.com":
		isEvent, _, err := CheckStackOverflowQuestionUpdate(linkURL, lastKnown)
		if err != nil {
			return false, err
		}

		return isEvent, nil
	default:
		return false, nil
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

func (s *Scrapper) AddLink(id int64, link domain.Link) error {
	slog.Info("Adding link", "link", link)

	if !validLink(link.URL) {
		slog.Error("Invalid link URL", "link", link)
		return ErrWrongURL{}
	}

	err := s.db.AddLink(id, link)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
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
