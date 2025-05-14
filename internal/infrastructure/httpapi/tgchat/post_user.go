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
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "400",
			"Invalid or missing tgID", err.Error(), "INVALID_TG_ID")

		return
	}

	err = h.UserAdder.AddUser(r.Context(), chatID)
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "500",
			"Failed to create user", err.Error(), "CREATE_USER_FAILED")

		return
	}

	w.WriteHeader(http.StatusOK)
}
