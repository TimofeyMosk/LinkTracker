package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/server"
)

func main() {
	config, err := application.ReadYAMLConfig("config.yaml")
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		return
	}

	application.InitLogger(config.BotConfig.LogsPath)

	wg := sync.WaitGroup{}

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

	bot := application.NewBot(scrapperHTTPClient, tgClient, 20)

	wg.Add(1)

	go func() {
		bot.Start()
		wg.Done()
	}()

	serv := server.InitServer(config.BotConfig.Address,
		server.InitBotRouting(bot),
		config.BotConfig.ReadTimeout,
		config.BotConfig.WriteTimeout)

	go application.StopBotSignalReceiving(bot, serv)

	if err := serv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Println("server failed to start or finished with error", err)
		slog.Error("server failed to start or finished with error", "error", err)
	} else {
		fmt.Println("server stopped")
		slog.Info("server stopped")
	}

	wg.Wait()
}
