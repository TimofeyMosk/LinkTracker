package links

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"LinkTracker/internal/infrastructure/dto"

	"LinkTracker/internal/domain"
	scrapperdto "LinkTracker/internal/infrastructure/dto/dto_scrapper"
	"LinkTracker/internal/infrastructure/httpapi"
)

type LinkDeleter interface {
	DeleteLink(ctx context.Context, id int64, link *domain.Link) (domain.Link, error)
}

type DeleteLinksHandler struct {
	LinkDeleter LinkDeleter
}

func (h DeleteLinksHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tgID, err := httpapi.GetTgIDFromString(r.Header.Get("Tg-Chat-Id"))
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "400",
			"Invalid or missing tgID", err.Error(), "INVALID_TG_ID")

		return
	}

	var removeLinkRequest scrapperdto.RemoveLinkRequest
	if err = json.NewDecoder(r.Body).Decode(&removeLinkRequest); err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "400",
			"Invalid or missing request body", err.Error(), "INVALID_REQUEST_BODY")

		return
	}

	link, err := dto.RemoveLinkRequestDTOToLink(removeLinkRequest)
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "400",
			"Missing required fields", err.Error(), "INVALID_REQUEST_BODY")

		return
	}

	deletedLink, err := h.LinkDeleter.DeleteLink(r.Context(), tgID, &link)
	if err != nil {
		if errors.As(err, &domain.ErrLinkNotExist{}) {
			httpapi.SendErrorResponse(w, http.StatusNotFound, "404",
				"Link not found", err.Error(), "LINK_NOT_EXIST")

			return
		}

		httpapi.SendErrorResponse(w, http.StatusBadRequest, "500",
			"Failed to delete link", err.Error(), "DELETE_LINK_FAILED")

		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	linkResponse := dto.LinkToLinkResponseDTO(&deletedLink)

	err = json.NewEncoder(w).Encode(linkResponse)
	if err != nil {
		slog.Error(err.Error())
	}
}
