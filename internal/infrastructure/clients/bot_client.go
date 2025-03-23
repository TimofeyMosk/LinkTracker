package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"LinkTracker/internal/domain"
	botdto "LinkTracker/internal/infrastructure/dto/dto_bot"
)

type BotHTTPClient struct {
	client     *http.Client
	botBaseURL *url.URL
}

func NewBotHTTPClient(botBaseURL string, timeout time.Duration) (*BotHTTPClient, error) {
	parsedURL, err := url.Parse(botBaseURL)
	if err != nil {
		return nil, err
	}

	return &BotHTTPClient{
		client:     &http.Client{Timeout: timeout},
		botBaseURL: parsedURL}, nil
}

func (c *BotHTTPClient) PostUpdates(link domain.Link, tgID int64) error {
	endpoint := c.botBaseURL.JoinPath("/updates")

	linkUpdate := botdto.LinkUpdate{
		Description: nil,
		Id:          &link.ID,
		TgChatIds:   &[]int64{tgID},
		Url:         &link.URL,
	}

	payload, err := json.Marshal(linkUpdate)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, endpoint.String(), bytes.NewReader(payload))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := c.client.Do(request)
	if err != nil {
		return err
	}

	defer func() {
		if Cerr := response.Body.Close(); Cerr != nil {
			slog.Error("could not close resource", "error", Cerr.Error())
		}
	}()

	switch response.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return HandleAPIErrorResponseFromBot(response)
	default:
		return domain.ErrUnexpectedStatusCode{StatusCode: response.StatusCode}
	}
}

func HandleAPIErrorResponseFromBot(resp *http.Response) error {
	var errorResponse botdto.ApiErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
		return fmt.Errorf("decode error response: %w", err)
	}

	var apiError = domain.ErrAPI{}
	if errorResponse.Code != nil {
		apiError.Code = *errorResponse.Code
	}

	if errorResponse.Description != nil {
		apiError.Description = *errorResponse.Description
	}

	if errorResponse.ExceptionMessage != nil {
		apiError.ExceptionMessage = *errorResponse.ExceptionMessage
	}

	if errorResponse.ExceptionName != nil {
		apiError.ExceptionName = *errorResponse.ExceptionName
	}

	if errorResponse.Stacktrace != nil {
		apiError.Stacktrace = *errorResponse.Stacktrace
	}

	return apiError
}
