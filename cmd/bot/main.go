package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"LinkTracker/internal/application/bot"

	"LinkTracker/internal/application"
	"LinkTracker/internal/infrastructure/clients"
	"LinkTracker/internal/infrastructure/server"
	"LinkTracker/pkg"
)

func main() {
	config, err := application.ReadYAMLConfig("config.yaml")
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		return
	}

	pkg.InitLogger(config.BotConfig.LogsPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scrapperHTTPClient, err := clients.NewScrapperHTTPClient(config.BotConfig.ScrapperBaseURL, config.BotConfig.ScrapperClientTimeout)
	if err != nil {
		fmt.Printf("Error creating scrapper client: %v\n", err)
		return
	}

	tgClient, err := clients.NewTelegramHTTPClient(config.BotConfig.TgToken)
	b := bot.NewBot(scrapperHTTPClient, tgClient)
	wg := sync.WaitGroup{}

	wg.Add(1)

	go func() {
		defer wg.Done()
		b.Run(ctx)
	}()

	if err != nil {
		fmt.Printf("Error creating tgClient: %v\n", err)
		return
	}

	serv := server.InitServer(config.BotConfig.Address,
		server.InitBotRouting(tgClient),
		config.BotConfig.ReadTimeout,
		config.BotConfig.WriteTimeout)

	wg.Add(1)

	go func() {
		defer wg.Done()
		application.StopBotSignalReceiving(cancel, serv)
	}()

	if err := serv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server failed to start or finished with error", "error", err)
	} else {
		slog.Info("server stopped")
	}

	wg.Wait()
}
