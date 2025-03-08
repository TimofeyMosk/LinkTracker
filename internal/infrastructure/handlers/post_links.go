package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	scrapperdto "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/dto/dto_scrapper"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
)

type PostLinkHandler struct{ Scrapper *application.Scrapper }

func (h PostLinkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tgChatID, err := getIDFromString(r.Header.Get("Tg-Chat-Id"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "INVALID_CHAT_ID",
			"Invalid or missing chat ID", err.Error(), "BadRequest")

		return
	}

	var addLinkRequest scrapperdto.AddLinkRequest
	if err = json.NewDecoder(r.Body).Decode(&addLinkRequest); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST_BODY",
			"Invalid or missing request body", err.Error(), "BadRequest")

		return
	}

	link, err := AddLinkRequestDtoToLink(addLinkRequest)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST_BODY",
			"Invalid or missing request body", err.Error(), "BadRequest")

		return
	}

	link, err = h.Scrapper.AddLink(tgChatID, link)
	if err != nil {
		if errors.As(err, &domain.ErrUserNotExist{}) {
			sendErrorResponse(w, http.StatusBadRequest, "CHAT_NOT_EXIST",
				"Chat not exist", err.Error(), "BadRequest")
		} else {
			sendErrorResponse(w, http.StatusBadRequest, "ADD_LINK_FAILED",
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
