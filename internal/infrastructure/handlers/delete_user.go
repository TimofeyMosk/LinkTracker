package handlers

import (
	"errors"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application"
)

type DeleteUserHandler struct{ Srapper *application.Scrapper }

func (h DeleteUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chatID, err := getIDFromString(r.PathValue("id"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "INVALID_CHAT_ID",
			"Invalid or missing chat ID", err.Error(), "BadRequest")

		return
	}

	err = h.Srapper.DeleteUser(chatID)
	if err != nil {
		if errors.As(err, &domain.ErrUserNotExist{}) {
			sendErrorResponse(w, http.StatusNotFound, "CHAT_NOT_EXIST",
				"Chat not exist", err.Error(), "Not Found")
		} else {
			sendErrorResponse(w, http.StatusBadRequest, "CHAT_NOT_DELETED",
				"Chat has not been deleted", err.Error(), "BadRequest")
		}

		return
	}

	w.WriteHeader(http.StatusOK)
}
