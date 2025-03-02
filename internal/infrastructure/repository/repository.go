package repository

import (
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"sync"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
)

type Repository struct {
	Links map[int64][]domain.Link
	mu    sync.RWMutex
}

func NewRepository() *Repository {
	slog.Info("Creating new repository")

	return &Repository{
		Links: make(map[int64][]domain.Link),
	}
}

func (r *Repository) GetAllUsers() ([]int64, error) {
	usersIDs := slices.Collect(maps.Keys(r.Links))
	return usersIDs, nil
}

func (r *Repository) GetAllLinks() ([]domain.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	set := make(map[string]bool)
	answer := make([]domain.Link, 0, len(r.Links))

	for _, value := range r.Links {
		for _, link := range value {
			if _, ok := set[link.URL]; !ok {
				answer = append(answer, link)
				set[link.URL] = true
			}
		}
	}

	return answer, nil
}

func (r *Repository) CreateUser(id int64) error {
	if r.UserExist(id) {
		return fmt.Errorf("user with id %d already exists", id)
	}

	r.Links[id] = []domain.Link{}

	return nil
}

func (r *Repository) DeleteUser(id int64) error {
	if !r.UserExist(id) {
		return application.ErrUserNotExist{}
	}

	delete(r.Links, id)

	return nil
}

func (r *Repository) GetLinks(id int64) ([]domain.Link, error) {
	if !r.UserExist(id) {
		return nil, application.ErrUserNotExist{}
	}

	return r.Links[id], nil
}

func (r *Repository) AddLink(id int64, link domain.Link) error {
	if !r.UserExist(id) {
		return application.ErrUserNotExist{}
	}

	r.mu.Lock()
	r.Links[id] = append(r.Links[id], link)
	r.mu.Unlock()

	return nil
}

func (r *Repository) DeleteLink(id int64, link domain.Link) (domain.Link, error) {
	if !r.UserExist(id) {
		return domain.Link{}, application.ErrUserNotExist{}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for index := range r.Links[id] {
		if r.Links[id][index].URL == link.URL {
			deletedLink := r.Links[id][index]
			r.Links[id] = append(r.Links[id][:index], r.Links[id][index+1:]...)

			return deletedLink, nil
		}
	}

	return domain.Link{}, fmt.Errorf("link with id %d not found", id)
}

func (r *Repository) UserExist(id int64) bool {
	r.mu.RLock()
	_, ok := r.Links[id]
	r.mu.RUnlock()

	return ok
}
