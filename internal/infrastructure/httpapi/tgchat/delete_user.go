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
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "400",
			"Invalid or missing tgID", err.Error(), "INVALID_TG_ID")

		return
	}

	err = h.UserDeleter.DeleteUser(r.Context(), chatID)
	if err != nil {
		if errors.As(err, &domain.ErrUserNotExist{}) {
			httpapi.SendErrorResponse(w, http.StatusNotFound, "404",
				"User not exist", err.Error(), "USER_NOT_EXIST")
		} else {
			httpapi.SendErrorResponse(w, http.StatusBadRequest, "500",
				"Chat has not been deleted", err.Error(), "CHAT_NOT_DELETED")
		}

		return
	}

	w.WriteHeader(http.StatusOK)
}
