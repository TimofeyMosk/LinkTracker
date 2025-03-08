package handlers

import (
	"net/http"
)

type PostUserHandler struct{ Scrapper Scrapper }

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
