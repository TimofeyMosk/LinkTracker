package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

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

	scrapperHTTPClient, err := clients.NewScrapperHTTPClient(config.BotConfig.ScrapperBaseURL, config.BotConfig.ScrapperClientTimeout)
	if err != nil {
		fmt.Printf("Error creating scrapper client: %v\n", err)
		return
	}

	bot := application.NewBot(scrapperHTTPClient)

	tgClient, err := clients.NewTelegramHTTPClient(config.BotConfig.TgToken, bot)
	if err != nil {
		fmt.Printf("Error creating tgClient: %v\n", err)
		return
	}

	go tgClient.Run()

	serv := server.InitServer(config.BotConfig.Address,
		server.InitBotRouting(tgClient),
		config.BotConfig.ReadTimeout,
		config.BotConfig.WriteTimeout)

	go application.StopBotSignalReceiving(tgClient, serv)

	if err := serv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Println("server failed to start or finished with error", err)
		slog.Error("server failed to start or finished with error", "error", err)
	} else {
		fmt.Println("server stopped")
		slog.Info("server stopped")
	}
}
