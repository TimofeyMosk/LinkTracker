package states

import (
	"context"
	"encoding/json"
	"net/http"

	scrapperdto "LinkTracker/internal/infrastructure/dto/dto_scrapper"
	"LinkTracker/internal/infrastructure/httpapi"
)

type StateCreator interface {
	CreateState(ctx context.Context, tgID int64, state int) error
}

type PostStatesHandler struct {
	StateCreator StateCreator
}

func (h PostStatesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	if stateRequest.State == nil {
		httpapi.SendErrorResponse(w, http.StatusInternalServerError, "400",
			"Missing required fields", "stateRequest.State == nil", "INVALID_REQUEST_BODY")
		return
	}

	err = h.StateCreator.CreateState(r.Context(), tgID, *stateRequest.State)
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusInternalServerError, "500",
			"Failed to create state", err.Error(), "CREATE_STATE_FAILED")

		return
	}

	w.WriteHeader(http.StatusOK)
}
