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

type LinkAdder interface {
	AddLink(ctx context.Context, id int64, link *domain.Link) (domain.Link, error)
}

type PostLinksHandler struct {
	LinkAdder LinkAdder
}

func (h PostLinksHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	link, err = h.LinkAdder.AddLink(r.Context(), tgID, &link)
	if err != nil {
		switch {
		case errors.As(err, &domain.ErrUserNotExist{}):
			httpapi.SendErrorResponse(w, http.StatusBadRequest, "404",
				"User not exist", err.Error(), "USER_NOT_EXIST")
		case errors.As(err, &domain.ErrLinkAlreadyTracking{}):
			httpapi.SendErrorResponse(w, http.StatusBadRequest, "400",
				"Links already tracking", err.Error(), "LINKS_ALREADY_EXIST")
		default:
			httpapi.SendErrorResponse(w, http.StatusBadRequest, "500",
				"Failed to added link", err.Error(), "ADD_LINK_FAILED")
		}

		return
	}

	linkResponse := scrapperdto.LinkResponse{
		Id:      &link.ID,
		Url:     &link.URL,
		Tags:    &link.Tags,
		Filters: &link.Filters,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(linkResponse)
	if err != nil {
		slog.Error(err.Error())
	}
}
