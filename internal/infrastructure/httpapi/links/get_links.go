package links

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"LinkTracker/internal/infrastructure/dto"

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
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "400",
			"Invalid or missing tgID", err.Error(), "INVALID_TG_ID")

		return
	}

	links, err := h.LinkGetter.GetUserLinks(r.Context(), tgChatID)
	if err != nil {
		if errors.As(err, &domain.ErrUserNotExist{}) {
			httpapi.SendErrorResponse(w, http.StatusBadRequest, "404",
				"User not exist", err.Error(), "USER_NOT_EXIST")

			return
		}

		httpapi.SendErrorResponse(w, http.StatusBadRequest, "500",
			"Links not received", err.Error(), "LINKS_NOT_RECEIVED")

		return
	}

	responseData := dto.LinksToListLinksResponseDTO(links)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(responseData)
	if err != nil {
		slog.Error(err.Error())
	}
}
