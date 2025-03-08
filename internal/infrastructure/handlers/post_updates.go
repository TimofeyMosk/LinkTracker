package handlers

import (
	"encoding/json"
	"net/http"

	botdto "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/dto/dto_bot"
)

type PostUpdatesHandler struct {
	Bot Bot
}

func (handler PostUpdatesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var requestBody botdto.LinkUpdate
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST_BODY",
			"Invalid or missing request body", err.Error(), "BadRequest")

		return
	}

	if requestBody.TgChatIds == nil || requestBody.Url == nil {
		sendErrorResponse(w, http.StatusBadRequest, "MISSING_REQUIRED_FIELDS",
			"`TgChatIds` or `Url` is missing", "", "BadRequest")
		return
	}

	for i := range *requestBody.TgChatIds {
		handler.Bot.SendMessage((*requestBody.TgChatIds)[i], "Было обновление : "+*requestBody.Url)
	}

	w.WriteHeader(http.StatusOK)
}
