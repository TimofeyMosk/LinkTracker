package main

import (
	"fmt"

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

	bot := application.NewBot(scrapperHTTPClient, tgClient)

	go func() {
		bot.Start()
	}()

	serv := server.InitServer(config.BotConfig.Addr,
		server.InitBotRouting(bot),
		config.BotConfig.ReadTimeout,
		config.BotConfig.WriteTimeout)

	if err := serv.ListenAndServe(); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
