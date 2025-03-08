package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
)

type GetLinksHandler struct{ Scrapper *application.Scrapper }

func (h GetLinksHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tgChatID, err := getIDFromString(r.Header.Get("Tg-Chat-Id"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "INVALID_CHAT_ID",
			"Invalid or missing chat ID", err.Error(), "BadRequest")

		return
	}

	links, err := h.Scrapper.GetLinks(tgChatID)
	if err != nil {
		if errors.As(err, &domain.ErrUserNotExist{}) {
			sendErrorResponse(w, http.StatusBadRequest, "CHAT_NOT_EXIST",
				"Chat not exist", err.Error(), "BadRequest")

			return
		}

		sendErrorResponse(w, http.StatusBadRequest, "LINKS_NOT_RECEIVED",
			"Links not received", err.Error(), "Server Error")

		return
	}

	responseData := domainLinksToDTO(links)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(responseData)
	if err != nil {
		slog.Error(err.Error())
	}
}
