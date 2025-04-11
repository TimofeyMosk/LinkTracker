package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"LinkTracker/internal/application"
	"LinkTracker/internal/application/bot"
	"LinkTracker/internal/infrastructure/clients"
	"LinkTracker/internal/infrastructure/server"
)

func main() {
	config, err := application.ReadYAMLConfig()
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scrapperHTTPClient, err := clients.NewScrapperHTTPClient(config.BotConfig.ScrapperBaseURL, config.BotConfig.ScrapperClientTimeout)
	if err != nil {
		fmt.Printf("Error creating scrapper client: %v\n", err)
		return
	}

	tgClient, err := clients.NewTelegramHTTPClient(config.BotConfig.TgToken)
	if err != nil {
		fmt.Printf("Error creating tgClient: %v\n", err)
		return
	}

	Bot := bot.NewBot(scrapperHTTPClient, tgClient)
	serv := server.InitServer(config.BotConfig.Address,
		server.InitBotRouting(Bot),
		config.BotConfig.ReadTimeout,
		config.BotConfig.WriteTimeout)

	wg := &sync.WaitGroup{}

	wg.Add(1)

	go func() {
		defer wg.Done()
		application.StopBotSignalReceiving(ctx, cancel, serv)
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()
		Bot.Run(ctx)
	}()

	if err := serv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server failed to start or finished with error", "error", err)
	} else {
		slog.Info("server stopped")
	}

	wg.Wait()
}
