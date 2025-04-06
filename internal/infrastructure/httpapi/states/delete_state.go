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
		httpapi.SendErrorResponse(w, http.StatusInternalServerError, "INVALID_CHAT_ID",
			"Invalid or missing chat ID", err.Error(), "BadRequest")

		return
	}

	err = h.StateDeleter.DeleteState(r.Context(), tgID)
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusInternalServerError, "FAILED",
			"Failed to delete state", err.Error(), "Server Error")

		return
	}

	w.WriteHeader(http.StatusOK)
}
