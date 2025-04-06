package tgchat

import (
	"context"
	"errors"
	"net/http"

	"LinkTracker/internal/domain"
	"LinkTracker/internal/infrastructure/httpapi"
)

type UserDeleter interface {
	DeleteUser(ctx context.Context, tgID int64) error
}

type DeleteUserHandler struct {
	UserDeleter UserDeleter
}

func (h DeleteUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chatID, err := httpapi.GetTgIDFromString(r.PathValue("id"))
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "INVALID_CHAT_ID",
			"Invalid or missing chat ID", err.Error(), "BadRequest")

		return
	}

	err = h.UserDeleter.DeleteUser(r.Context(), chatID)
	if err != nil {
		if errors.As(err, &domain.ErrUserNotExist{}) {
			httpapi.SendErrorResponse(w, http.StatusNotFound, "CHAT_NOT_EXIST",
				"Chat not exist", err.Error(), "Not Found")
		} else {
			httpapi.SendErrorResponse(w, http.StatusBadRequest, "CHAT_NOT_DELETED",
				"Chat has not been deleted", err.Error(), "BadRequest")
		}

		return
	}

	w.WriteHeader(http.StatusOK)
}
