package repository

import (
	"fmt"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"sync"
)

type Repository struct {
	Links map[int64][]domain.Link
	mu    sync.RWMutex
}

func NewRepository() *Repository {
	return &Repository{
		Links: make(map[int64][]domain.Link),
	}
}
func (r *Repository) CreateUser(id int64) error {
	if r.userExist(id) {
		return fmt.Errorf("user with id %d already exists", id)
	}

	r.Links[id] = []domain.Link{}

	return nil
}

func (r *Repository) DeleteUser(id int64) error {
	if !r.userExist(id) {
		return fmt.Errorf("user with id %d not exists", id)
	}

	delete(r.Links, id)

	return nil
}

func (r *Repository) GetLinks(id int64) ([]domain.Link, error) {
	if !r.userExist(id) {
		return nil, fmt.Errorf("user with id %d not exists", id)
	}

	return r.Links[id], nil
}

func (r *Repository) AddLink(id int64, link domain.Link) error {
	if !r.userExist(id) {
		return fmt.Errorf("user with id %d not exists", id)
	}

	if r.linkExist(id, link.Url) {
		return fmt.Errorf("link with id %d already exists", id)
	}

	r.mu.Lock()
	r.Links[id] = append(r.Links[id], link)
	r.mu.Unlock()

	return nil
}

func (r *Repository) DeleteLink(id int64, link domain.Link) error {
	if !r.userExist(id) {
		return fmt.Errorf("user with id %d not exists", id)
	}

	if !r.linkExist(id, link.Url) {
		return fmt.Errorf("link with id %d not exists", id)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for index := range r.Links[id] {
		if r.Links[id][index].Url == link.Url {
			r.Links[id] = append(r.Links[id][:index], r.Links[id][index+1:]...)
			return nil
		}
	}

	return fmt.Errorf("link with id %d not found", id)
}

func (r *Repository) userExist(id int64) bool {
	r.mu.RLock()
	_, ok := r.Links[id]
	r.mu.RUnlock()

	return ok
}

func (r *Repository) linkExist(id int64, link string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for index := range r.Links[id] {
		if r.Links[id][index].Url == link {
			return true
		}
	}

	return false
}
