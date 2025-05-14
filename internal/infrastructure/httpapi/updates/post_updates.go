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
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "400",
			"Invalid or missing request body", err.Error(), "INVALID_REQUEST_BODY")
		return
	}

	if requestBody.TgChatIds == nil || requestBody.Url == nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "400",
			"\"TgChatIds\" or \"Url\" is missing", "requestBody.TgChatIds == nil || requestBody.Url == nil",
			"MISSING_REQUIRED_FIELDS")

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
