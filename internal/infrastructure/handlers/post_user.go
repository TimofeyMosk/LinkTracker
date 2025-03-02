package handlers

import (
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application"
)

type PostUserHandler struct{ Scrapper *application.Scrapper }

func (h PostUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chatID, err := getIDFromString(r.PathValue("id"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "INVALID_CHAT_ID",
			"Invalid or missing chat ID", err.Error(), "BadRequest")

		return
	}

	err = h.Scrapper.AddUser(chatID)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "CREATE_CHAT_FAILED",
			"Failed to create chat", err.Error(), "BadRequest")

		return
	}

	w.WriteHeader(http.StatusOK)
}
