package updates

import (
	"encoding/json"
	"net/http"

	botdto "LinkTracker/internal/infrastructure/dto/dto_bot"
	"LinkTracker/internal/infrastructure/httpapi"
)

type MessageSender interface {
	SendMessage(chatID int64, message string)
}

type PostUpdatesHandler struct {
	MessageSender MessageSender
}

func (handler PostUpdatesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var requestBody botdto.LinkUpdate
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST_BODY",
			"Invalid or missing request body", err.Error(), "BadRequest")

		return
	}

	if requestBody.TgChatIds == nil || requestBody.Url == nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "MISSING_REQUIRED_FIELDS",
			"\"TgChatIds\" or \"Url\" is missing", "", "BadRequest")
		return
	}

	for i := range *requestBody.TgChatIds {
		handler.MessageSender.SendMessage((*requestBody.TgChatIds)[i], "Было обновление : "+*requestBody.Url)
	}

	w.WriteHeader(http.StatusOK)
}
