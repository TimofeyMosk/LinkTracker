package links

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"LinkTracker/internal/domain"
	"LinkTracker/internal/infrastructure/httpapi"
)

type LinkGetter interface {
	GetUserLinks(ctx context.Context, tgID int64) ([]domain.Link, error)
}

type GetLinksHandler struct {
	LinkGetter LinkGetter
}

func (h GetLinksHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tgChatID, err := httpapi.GetTgIDFromString(r.Header.Get("Tg-Chat-Id"))
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "INVALID_CHAT_ID",
			"Invalid or missing chat ID", err.Error(), "BadRequest")

		return
	}

	links, err := h.LinkGetter.GetUserLinks(r.Context(), tgChatID)
	if err != nil {
		if errors.As(err, &domain.ErrUserNotExist{}) {
			httpapi.SendErrorResponse(w, http.StatusBadRequest, "CHAT_NOT_EXIST",
				"Chat not exist", err.Error(), "BadRequest")

			return
		}

		httpapi.SendErrorResponse(w, http.StatusBadRequest, "LINKS_NOT_RECEIVED",
			"Links not received", err.Error(), "Server Error")

		return
	}

	responseData := httpapi.LinksToDTO(links)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(responseData)
	if err != nil {
		slog.Error(err.Error())
	}
}
