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

type LinkDeleter interface {
	DeleteLink(ctx context.Context, id int64, link *domain.Link) (domain.Link, error)
}

type DeleteLinksHandler struct {
	LinkDeleter LinkDeleter
}

func (h DeleteLinksHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tgChatID, err := httpapi.GetTgIDFromString(r.Header.Get("Tg-Chat-Id"))
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "INVALID_CHAT_ID",
			"Invalid or missing chat ID", err.Error(), "BadRequest")

		return
	}

	var removeLinkRequest scrapperdto.RemoveLinkRequest
	if err = json.NewDecoder(r.Body).Decode(&removeLinkRequest); err != nil {
		httpapi.SendErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST_BODY",
			"Invalid or missing request body", err.Error(), "BadRequest")

		return
	}

	link := domain.Link{URL: *removeLinkRequest.Link, Tags: nil, Filters: nil, ID: 0}

	deletedLink, err := h.LinkDeleter.DeleteLink(r.Context(), tgChatID, &link)
	if err != nil {
		if errors.As(err, &domain.ErrUserNotExist{}) {
			httpapi.SendErrorResponse(w, http.StatusNotFound, "CHAT_NOT_EXIST",
				"Chat not exist", err.Error(), "Not Found")

			return
		}

		httpapi.SendErrorResponse(w, http.StatusBadRequest, "DELETE_LINK_FAILED",
			"Failed to delete link", err.Error(), "BadRequest")

		return
	}

	linkResponse := scrapperdto.LinkResponse{
		Url:     &deletedLink.URL,
		Id:      &tgChatID,
		Tags:    &deletedLink.Tags,
		Filters: &deletedLink.Filters,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(linkResponse)
	if err != nil {
		slog.Error(err.Error())
	}
}
