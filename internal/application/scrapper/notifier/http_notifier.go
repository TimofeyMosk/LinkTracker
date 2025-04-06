package notifier

import (
	"context"

	"LinkTracker/internal/domain"
)

type BotClient interface {
	PostUpdates(ctx context.Context, link *domain.Link, tgID []int64, description string) error
}

type HTTPNotifier struct {
	botClient BotClient
}

func NewHTTPNotifier(botClient BotClient) *HTTPNotifier {
	return &HTTPNotifier{botClient: botClient}
}

func (n HTTPNotifier) PostUpdates(ctx context.Context, link *domain.Link, tgID []int64, description string) error {
	err := n.botClient.PostUpdates(ctx, link, tgID, description)
	if err != nil {
		return err
	}

	return nil
}
