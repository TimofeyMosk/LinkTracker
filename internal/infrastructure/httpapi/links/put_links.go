package links

import (
	"context"
	"encoding/json"
	"net/http"

	"LinkTracker/internal/domain"
	scrapperdto "LinkTracker/internal/infrastructure/dto/dto_scrapper"
	"LinkTracker/internal/infrastructure/httpapi"
)

type LinkUpdater interface {
	UpdateLink(ctx context.Context, tgID int64, link *domain.Link) error
}

type PutLinksHandler struct {
	LinkUpdater LinkUpdater
}

func (h PutLinksHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tgID, err := httpapi.GetTgIDFromString(r.Header.Get("Tg-Chat-Id"))
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "INVALID_CHAT_ID",
			"Invalid or missing chat ID", err.Error(), "BadRequest")

		return
	}

	var LinkRequest scrapperdto.LinkRequest
	if err = json.NewDecoder(r.Body).Decode(&LinkRequest); err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST_BODY",
			"Invalid or missing request body", err.Error(), "BadRequest")

		return
	}

	link, err := httpapi.LinkRequestDtoToLink(LinkRequest)
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST_BODY",
			"Invalid or missing required fields", err.Error(), "BadRequest")

		return
	}

	err = h.LinkUpdater.UpdateLink(r.Context(), tgID, &link)
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "UPDATE_LINK_FAILED",
			"Failed to update link", err.Error(), "Server Error")

		return
	}

	w.WriteHeader(http.StatusOK)
}
