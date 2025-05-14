package states

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"LinkTracker/internal/domain"
	scrapperdto "LinkTracker/internal/infrastructure/dto/dto_scrapper"
	"LinkTracker/internal/infrastructure/httpapi"
)

type StateGetter interface {
	GetState(ctx context.Context, tgID int64) (int, domain.Link, error)
}

type GetStatesHandler struct {
	StateGetter StateGetter
}

func (h GetStatesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tgID, err := httpapi.GetTgIDFromString(r.Header.Get("Tg-Chat-Id"))
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusInternalServerError, "400",
			"Invalid or missing tgID", err.Error(), "INVALID_TG_ID")

		return
	}

	state, link, err := h.StateGetter.GetState(r.Context(), tgID)
	if err != nil {
		httpapi.SendErrorResponse(w, http.StatusInternalServerError, "500",
			"Failed to get state", err.Error(), "GET_STATE_FAILED")

		return
	}

	var responseData scrapperdto.StateResponse
	responseData.State = &state
	responseData.Link = &link.URL
	responseData.Tags = &link.Tags
	responseData.Filters = &link.Filters

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(responseData)
	if err != nil {
		slog.Error(err.Error())
	}
}
