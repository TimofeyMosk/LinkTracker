package tgchat

import (
	"net/http"

	"LinkTracker/internal/infrastructure/httpapi"
)

type UserAdder interface {
	AddUser(int64) error
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

	err = h.UserAdder.AddUser(chatID)
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "CREATE_CHAT_FAILED",
			"Failed to create chat", err.Error(), "BadRequest")

		return
	}

	w.WriteHeader(http.StatusOK)
}
