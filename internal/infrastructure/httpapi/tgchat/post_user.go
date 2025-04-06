package tgchat

import (
	"context"
	"net/http"

	"LinkTracker/internal/infrastructure/httpapi"
)

type UserAdder interface {
	AddUser(ctx context.Context, tgID int64) error
}

type PostUserHandler struct {
	UserAdder UserAdder
}

func (h PostUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chatID, err := httpapi.GetTgIDFromString(r.PathValue("id"))
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "INVALID_CHAT_ID",
			"Invalid or missing chat ID", err.Error(), "BadRequest")

		return
	}

	err = h.UserAdder.AddUser(r.Context(), chatID)
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "CREATE_CHAT_FAILED",
			"Failed to create chat", err.Error(), "BadRequest")

		return
	}

	w.WriteHeader(http.StatusOK)
}
