package updates

import (
	"context"
	"encoding/json"
	"net/http"

	botdto "LinkTracker/internal/infrastructure/dto/dto_bot"
	"LinkTracker/internal/infrastructure/httpapi"
)

type UpdateSender interface {
	UpdateSend(ctx context.Context, tgIDs []int64, url string, description string)
}

type PostUpdatesHandler struct {
	UpdateSender UpdateSender
}

func (h PostUpdatesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	if requestBody.Description == nil {
		requestBody.Description = strPtr("")
	}

	h.UpdateSender.UpdateSend(r.Context(), *requestBody.TgChatIds, *requestBody.Url, *requestBody.Description)

	w.WriteHeader(http.StatusOK)
}

func strPtr(s string) *string {
	return &s
}
