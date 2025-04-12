package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"LinkTracker/internal/infrastructure/dto"

	"LinkTracker/internal/domain"
	scrapperdto "LinkTracker/internal/infrastructure/dto/dto_scrapper"
)

type ScrapperHTTPClient struct {
	client          *http.Client
	scrapperBaseURL *url.URL
}

func NewScrapperHTTPClient(scrapperBaseURL string, timeout time.Duration) (*ScrapperHTTPClient, error) {
	parsedURL, err := url.Parse(scrapperBaseURL)
	if err != nil {
		return nil, err
	}

	return &ScrapperHTTPClient{
		client:          &http.Client{Timeout: timeout},
		scrapperBaseURL: parsedURL}, nil
}

func (c *ScrapperHTTPClient) RegisterUser(ctx context.Context, tgID int64) error {
	endpoint := c.scrapperBaseURL.JoinPath(fmt.Sprintf("/tg-chat/%d", tgID))

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), http.NoBody)
	if err != nil {
		return err
	}

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
		return HandleAPIErrorResponseFromScrapper(response)
	default:
		return domain.ErrUnexpectedStatusCode{StatusCode: response.StatusCode}
	}
}

func (c *ScrapperHTTPClient) DeleteUser(ctx context.Context, tgID int64) error {
	endpoint := c.scrapperBaseURL.JoinPath(fmt.Sprintf("/tg-chat/%d", tgID))

	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint.String(), http.NoBody)
	if err != nil {
		return err
	}

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
	case http.StatusNotFound:
		return HandleAPIErrorResponseFromScrapper(response)
	case http.StatusBadRequest:
		return HandleAPIErrorResponseFromScrapper(response)
	default:
		return domain.ErrUnexpectedStatusCode{StatusCode: response.StatusCode}
	}
}

func (c *ScrapperHTTPClient) GetLinks(ctx context.Context, tgID int64) ([]domain.Link, error) {
	endpoint := c.scrapperBaseURL.JoinPath("/links")

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), http.NoBody)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Tg-Chat-Id", fmt.Sprint(tgID))

	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}

	defer func() {
		if Cerr := response.Body.Close(); Cerr != nil {
			slog.Error("could not close resource", "error", Cerr.Error())
		}
	}()

	switch response.StatusCode {
	case http.StatusOK:
		var listLinksResponse scrapperdto.ListLinksResponse
		if err := json.NewDecoder(response.Body).Decode(&listLinksResponse); err != nil {
			return nil, err
		}

		if listLinksResponse.Links == nil || listLinksResponse.Size == nil {
			return []domain.Link{}, nil
		}

		listLinks := make([]domain.Link, 0, len(*listLinksResponse.Links))
		for _, link := range *listLinksResponse.Links {
			listLinks = append(listLinks, domain.Link{
				URL:     *link.Url,
				Tags:    *link.Tags,
				Filters: *link.Filters,
				ID:      *link.Id,
			})
		}

		return listLinks, nil
	case http.StatusBadRequest:
		return nil, HandleAPIErrorResponseFromScrapper(response)
	default:
		return nil, domain.ErrUnexpectedStatusCode{StatusCode: response.StatusCode}
	}
}

func (c *ScrapperHTTPClient) AddLink(ctx context.Context, tgID int64, link *domain.Link) error {
	endpoint := c.scrapperBaseURL.JoinPath("/links")

	payload, err := json.Marshal(dto.LinkToLinkRequestDTO(link))
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(payload))
	if err != nil {
		return err
	}

	request.Header.Set("Tg-Chat-Id", fmt.Sprint(tgID))

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
		var responseLink scrapperdto.LinkResponse
		if err := json.NewDecoder(response.Body).Decode(&responseLink); err != nil {
			slog.Error("Failed to decode response from scrapper", "error", err.Error())
			return nil
		}

		if responseLink.Url != nil && responseLink.Id != nil {
			slog.Info("Added link", "url", *responseLink.Url, "ID", *responseLink.Id)
		}

		return nil
	case http.StatusBadRequest:
		return HandleAPIErrorResponseFromScrapper(response)
	default:
		return domain.ErrUnexpectedStatusCode{StatusCode: response.StatusCode}
	}
}

func (c *ScrapperHTTPClient) RemoveLink(ctx context.Context, tgID int64, link *domain.Link) error {
	endpoint := c.scrapperBaseURL.JoinPath("/links")

	payload, err := json.Marshal(dto.LinkToRemoveListRequestDTO(link))
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint.String(), bytes.NewReader(payload))
	if err != nil {
		return err
	}

	request.Header.Set("Tg-Chat-Id", fmt.Sprint(tgID))

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
		var responseLink scrapperdto.LinkResponse
		if err := json.NewDecoder(response.Body).Decode(&responseLink); err != nil {
			slog.Error("Failed to decode response from scrapper", "error", err.Error())
		}

		if responseLink.Url != nil && responseLink.Id != nil {
			slog.Info("Removed link", "url", *responseLink.Url)
		}

		return nil
	case http.StatusBadRequest:
		return HandleAPIErrorResponseFromScrapper(response)
	case http.StatusNotFound:
		return HandleAPIErrorResponseFromScrapper(response)
	default:
		return domain.ErrUnexpectedStatusCode{StatusCode: response.StatusCode}
	}
}

func (c *ScrapperHTTPClient) UpdateLink(ctx context.Context, tgID int64, link *domain.Link) error {
	endpoint := c.scrapperBaseURL.JoinPath("/links")

	payload, err := json.Marshal(dto.LinkToLinkRequestDTO(link))
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint.String(), bytes.NewReader(payload))
	if err != nil {
		return err
	}

	request.Header.Set("Tg-Chat-Id", fmt.Sprint(tgID))

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
		return HandleAPIErrorResponseFromScrapper(response)
	default:
		return domain.ErrUnexpectedStatusCode{StatusCode: response.StatusCode}
	}
}

func (c *ScrapperHTTPClient) CreateState(ctx context.Context, tgID int64, state int) error {
	endpoint := c.scrapperBaseURL.JoinPath("/states")

	payload, err := json.Marshal(scrapperdto.StateRequest{State: &state})
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(payload))
	if err != nil {
		return err
	}

	request.Header.Set("Tg-Chat-Id", fmt.Sprint(tgID))

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
	case http.StatusInternalServerError:
		return HandleAPIErrorResponseFromScrapper(response)
	default:
		return domain.ErrUnexpectedStatusCode{StatusCode: response.StatusCode}
	}
}

func (c *ScrapperHTTPClient) DeleteState(ctx context.Context, tgID int64) error {
	endpoint := c.scrapperBaseURL.JoinPath("/states")

	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint.String(), http.NoBody)
	if err != nil {
		return err
	}

	request.Header.Set("Tg-Chat-Id", fmt.Sprint(tgID))

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
	case http.StatusInternalServerError:
		return HandleAPIErrorResponseFromScrapper(response)
	default:
		return domain.ErrUnexpectedStatusCode{StatusCode: response.StatusCode}
	}
}

func (c *ScrapperHTTPClient) UpdateState(ctx context.Context, tgID int64, state int, link *domain.Link) error {
	endpoint := c.scrapperBaseURL.JoinPath("/states")

	payload, err := json.Marshal(scrapperdto.StateRequest{State: &state, Link: &link.URL, Tags: &link.Tags, Filters: &link.Filters})
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint.String(), bytes.NewReader(payload))
	if err != nil {
		return err
	}

	request.Header.Set("Tg-Chat-Id", fmt.Sprint(tgID))

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
	case http.StatusInternalServerError:
		return HandleAPIErrorResponseFromScrapper(response)
	default:
		return domain.ErrUnexpectedStatusCode{StatusCode: response.StatusCode}
	}
}

func (c *ScrapperHTTPClient) GetState(ctx context.Context, tgID int64) (int, domain.Link, error) {
	endpoint := c.scrapperBaseURL.JoinPath("/states")

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), http.NoBody)
	if err != nil {
		return -1, domain.Link{}, err
	}

	request.Header.Set("Tg-Chat-Id", fmt.Sprint(tgID))

	response, err := c.client.Do(request)
	if err != nil {
		return -1, domain.Link{}, err
	}

	defer func() {
		if Cerr := response.Body.Close(); Cerr != nil {
			slog.Error("could not close resource", "error", Cerr.Error())
		}
	}()

	switch response.StatusCode {
	case http.StatusOK:
		var responseData scrapperdto.StateResponse
		if err := json.NewDecoder(response.Body).Decode(&responseData); err != nil {
			return -1, domain.Link{}, err
		}

		if responseData.Link == nil {
			*responseData.Link = ""
		}

		if responseData.Tags == nil {
			*responseData.Tags = []string{}
		}

		if responseData.Filters == nil {
			*responseData.Filters = []string{}
		}

		responseLink := domain.Link{URL: *responseData.Link, Tags: *responseData.Tags, Filters: *responseData.Filters}

		return *responseData.State, responseLink, nil
	case http.StatusInternalServerError:
		return -1, domain.Link{}, HandleAPIErrorResponseFromScrapper(response)
	default:
		return -1, domain.Link{}, domain.ErrUnexpectedStatusCode{StatusCode: response.StatusCode}
	}
}

func HandleAPIErrorResponseFromScrapper(resp *http.Response) error {
	var errorResponse scrapperdto.ApiErrorResponse
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
