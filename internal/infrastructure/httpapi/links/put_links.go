package links

import (
	"context"
	"encoding/json"
	"net/http"

	"LinkTracker/internal/infrastructure/dto"

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
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "400",
			"Invalid or missing tgID", err.Error(), "INVALID_TG_ID")

		return
	}

	var LinkRequest scrapperdto.LinkRequest
	if err = json.NewDecoder(r.Body).Decode(&LinkRequest); err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "400",
			"Invalid or missing request body", err.Error(), "INVALID_REQUEST_BODY")

		return
	}

	link, err := dto.LinkRequestDTOToLink(LinkRequest)
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "400",
			"Missing required fields", err.Error(), "INVALID_REQUEST_BODY")

		return
	}

	err = h.LinkUpdater.UpdateLink(r.Context(), tgID, &link)
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "500",
			"Failed to update link", err.Error(), "UPDATE_LINK_FAILED")

		return
	}

	w.WriteHeader(http.StatusOK)
}
