package states

import (
	"context"
	"net/http"

	"LinkTracker/internal/infrastructure/httpapi"
)

type StateDeleter interface {
	DeleteState(ctx context.Context, tgID int64) error
}

type DeleteStatesHandler struct {
	StateDeleter StateDeleter
}

func (h DeleteStatesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tgID, err := httpapi.GetTgIDFromString(r.Header.Get("Tg-Chat-Id"))
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusInternalServerError, "400",
			"Invalid or missing tgID", err.Error(), "INVALID_TG_ID")

		return
	}

	err = h.StateDeleter.DeleteState(r.Context(), tgID)
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusInternalServerError, "500",
			"Failed to delete state", err.Error(), "DELETE_STATE_FAILED")

		return
	}

	w.WriteHeader(http.StatusOK)
}
