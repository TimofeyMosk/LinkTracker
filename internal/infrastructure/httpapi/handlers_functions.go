package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"LinkTracker/internal/domain"
	scrapperdto "LinkTracker/internal/infrastructure/dto/dto_scrapper"
)

func SendErrorResponse(w http.ResponseWriter, statusCode int, code, description, exceptionMessage, exceptionName string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := scrapperdto.ApiErrorResponse{
		Code:             &code,
		Description:      &description,
		ExceptionMessage: &exceptionMessage,
		ExceptionName:    &exceptionName,
		Stacktrace:       nil,
	}

	err := json.NewEncoder(w).Encode(errorResponse)
	if err != nil {
		slog.Error(err.Error())
	}
}

func LinksToDTO(links []domain.Link) scrapperdto.ListLinksResponse {
	linksResponse := make([]scrapperdto.LinkResponse, len(links))
	for i := range links {
		linksResponse[i] = scrapperdto.LinkResponse{
			Id:      &links[i].ID,
			Url:     &links[i].URL,
			Tags:    &links[i].Tags,
			Filters: &links[i].Filters,
		}
	}

	length := int32(len(linksResponse)) //nolint:gosec //api contract compliance(+ overflow is unlikely to be possible in real life)

	return scrapperdto.ListLinksResponse{Links: &linksResponse, Size: &length}
}

func GetTgIDFromString(s string) (int64, error) {
	if s == "" {
		return 0, domain.ErrEmptyString{}
	}

	chatID, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return chatID, nil
}
