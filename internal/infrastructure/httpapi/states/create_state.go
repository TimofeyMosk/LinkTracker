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
		httpapi.SendErrorResponse(w, http.StatusInternalServerError, "INVALID_CHAT_ID",
			"Invalid or missing chat ID", err.Error(), "BadRequest")

		return
	}

	var stateRequest scrapperdto.StateRequest
	if err = json.NewDecoder(r.Body).Decode(&stateRequest); err != nil {
		httpapi.SendErrorResponse(w, http.StatusInternalServerError, "INVALID_REQUEST_BODY",
			"Invalid or missing request body", err.Error(), "BadRequest")

		return
	}

	err = h.StateCreator.CreateState(r.Context(), tgID, *stateRequest.State)
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusInternalServerError, "FAILED",
			"Failed to create state", err.Error(), "Server Error")

		return
	}

	w.WriteHeader(http.StatusOK)
}
