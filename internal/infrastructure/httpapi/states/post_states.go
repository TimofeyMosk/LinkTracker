package states

import (
	"context"
	"net/http"
)

type StateCreator interface {
	CreateState(ctx context.Context, tgID int64, state int) error
}

type PostUserHandler struct {
	StateCreator StateCreator
}

func (h *PostUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
