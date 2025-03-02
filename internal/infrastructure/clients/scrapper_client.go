package clients

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	botdto "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/dto/dto_bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/requests"
)

type ScrapperClient struct {
	botBaseURL string
}

func NewScrapperClient(botBaseURL string) *ScrapperClient {
	return &ScrapperClient{botBaseURL: botBaseURL}
}

func (s *ScrapperClient) PostUpdates(link domain.Link, tgID int64) error {
	linkUpdate := botdto.LinkUpdate{
		Description: nil,
		Id:          nil,
		TgChatIds:   &[]int64{tgID},
		Url:         &link.URL,
	}

	var data bytes.Buffer

	err := json.NewEncoder(&data).Encode(linkUpdate)
	if err != nil {
		return err
	}

	resp, err := requests.PostRequest(s.botBaseURL+"/updates", &data)
	defer func() {
		if Cerr := resp.Body.Close(); Cerr != nil {
			slog.Error(Cerr.Error())
		}
	}()

	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		slog.Info("Sending of messages was successful")
		return nil
	case http.StatusBadRequest:
		var errorResponce botdto.ApiErrorResponse

		err = json.NewDecoder(resp.Body).Decode(&errorResponce)
		if err != nil {
			slog.Error(err.Error())
			return err
		}

		slog.Error("Error PostUpdates", "response", errorResponce)

		return nil
	default:
		return nil
	}
}
