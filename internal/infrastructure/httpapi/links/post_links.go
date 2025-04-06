package links

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"LinkTracker/internal/domain"
	scrapperdto "LinkTracker/internal/infrastructure/dto/dto_scrapper"
	"LinkTracker/internal/infrastructure/httpapi"
)

type LinkAdder interface {
	AddLink(ctx context.Context, id int64, link *domain.Link) (domain.Link, error)
}

type PostLinkHandler struct {
	LinkAdder LinkAdder
}

func (h PostLinkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tgChatID, err := httpapi.GetTgIDFromString(r.Header.Get("Tg-Chat-Id"))
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "INVALID_CHAT_ID",
			"Invalid or missing chat ID", err.Error(), "BadRequest")

		return
	}

	var addLinkRequest scrapperdto.AddLinkRequest
	if err = json.NewDecoder(r.Body).Decode(&addLinkRequest); err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST_BODY",
			"Invalid or missing request body", err.Error(), "BadRequest")

		return
	}

	link, err := AddLinkRequestDtoToLink(addLinkRequest)
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST_BODY",
			"Invalid or missing request body", err.Error(), "BadRequest")

		return
	}

	link, err = h.LinkAdder.AddLink(r.Context(), tgChatID, &link)
	if err != nil {
		switch {
		case errors.As(err, &domain.ErrUserNotExist{}):
			httpapi.SendErrorResponse(w, http.StatusBadRequest, "CHAT_NOT_EXIST",
				"Chat not exist", err.Error(), "BadRequest")
		case errors.As(err, &domain.ErrLinkAlreadyTracking{}):
			httpapi.SendErrorResponse(w, http.StatusBadRequest, "LINKS_ALREADY_EXIST",
				"Links already tracking", err.Error(), "BadRequest")
		default:
			httpapi.SendErrorResponse(w, http.StatusBadRequest, "ADD_LINK_FAILED",
				"Failed to added link", err.Error(), "Server Error")
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

func AddLinkRequestDtoToLink(addLinkRequest scrapperdto.AddLinkRequest) (domain.Link, error) {
	var link domain.Link

	if addLinkRequest.Link == nil {
		return link, domain.ErrNoRequiredAttribute{Attribute: "link"}
	}

	link.URL = *addLinkRequest.Link
	link.Tags = *addLinkRequest.Tags
	link.Filters = *addLinkRequest.Filters

	return link, nil
}
