package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	botdto "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/dto/dto_bot"
)

type PostUpdatesHandler struct{}

func (handler PostUpdatesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var requestBody botdto.LinkUpdate
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST_BODY",
			"Invalid or missing request body", err.Error(), "BadRequest")
	}

	// logic tg bots
	fmt.Printf("%+v", requestBody)

	w.WriteHeader(http.StatusOK)
}
