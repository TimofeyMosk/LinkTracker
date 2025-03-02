package clients

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"

	scrapperdto "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/dto/dto_scrapper"

	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/requests"
)

type BotClient struct {
	scrapperBaseURL string
}

func NewBotClient(scrapperBaseURL string) *BotClient {
	return &BotClient{scrapperBaseURL: scrapperBaseURL}
}

func (c *BotClient) RegisterUser(tgID int64) error {
	url := fmt.Sprintf("%s/tg-chat/%d", c.scrapperBaseURL, tgID)
	resp, err := requests.PostRequest(url, nil)
	if err != nil {
		slog.Error("Post request failed", "error", err)
		return err
	}
	if resp != nil {
		defer func() {
			if cerr := resp.Body.Close(); cerr != nil {
				slog.Error("Failed to close response body", "error", cerr)
			}
		}()
	}

	if resp.StatusCode == http.StatusOK {
		slog.Info("RegisterUser was successful")
		return nil
	}

	var errorResponse scrapperdto.ApiErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
		slog.Error("Failed to decode error response", "error", err)
		return domain.ErrRegistrationUser{}
	}

	slog.Error("Error registering user", "response", errorResponse)
	return domain.ErrRegistrationUser{}
}

func (c *BotClient) DeleteUser(tgID int64) error {
	url := fmt.Sprintf("%s/tg-chat/%d", c.scrapperBaseURL, tgID)
	resp, err := requests.DeleteRequest(url, nil)
	if err != nil {
		slog.Error("Delete request failed", "error", err)
		return err
	}
	if resp != nil {
		defer func() {
			if cerr := resp.Body.Close(); cerr != nil {
				slog.Error("Failed to close response body", "error", cerr)
			}
		}()
	}

	if resp.StatusCode == http.StatusOK {
		slog.Info("DeleteUser was successful")
		return nil
	}

	var errorResponse scrapperdto.ApiErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
		slog.Error("Failed to decode error response", "error", err)
		return domain.ErrDeletionUser{}
	}

	slog.Error("Error Deleting user", "response", errorResponse)
	return domain.ErrDeletionUser{}
}

func (c *BotClient) AddLink() error {

}
