package states

import (
	"context"
	"encoding/json"
	"net/http"

	"LinkTracker/internal/domain"
	scrapperdto "LinkTracker/internal/infrastructure/dto/dto_scrapper"
	"LinkTracker/internal/infrastructure/httpapi"
)

type StateUpdater interface {
	UpdateState(ctx context.Context, tgID int64, state int, link *domain.Link) error
}

type PutStatesHandler struct {
	StateUpdater StateUpdater
}

func (h PutStatesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tgID, err := httpapi.GetTgIDFromString(r.Header.Get("Tg-Chat-Id"))
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusInternalServerError, "400",
			"Invalid or missing tgID", err.Error(), "INVALID_TG_ID")

		return
	}

	var stateRequest scrapperdto.StateRequest
	if err = json.NewDecoder(r.Body).Decode(&stateRequest); err != nil {
		httpapi.SendErrorResponse(w, http.StatusInternalServerError, "400",
			"Invalid or missing request body", err.Error(), "INVALID_REQUEST_BODY")

		return
	}

	if stateRequest.Tags == nil {
		stateRequest.Tags = &[]string{}
	}

	if stateRequest.Filters == nil {
		stateRequest.Filters = &[]string{}
	}

	link := domain.Link{URL: *stateRequest.Link, Tags: *stateRequest.Tags, Filters: *stateRequest.Filters}

	err = h.StateUpdater.UpdateState(r.Context(), tgID, *stateRequest.State, &link)
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusInternalServerError, "500",
			"Failed to update state", err.Error(), "UPDATE_STATE_ERROR")

		return
	}

	w.WriteHeader(http.StatusOK)
}
