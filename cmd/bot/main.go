package main

import (
	"fmt"

	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/server"
)

func main() {
	config, err := application.ReadYAMLConfig("config.yaml")
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		return
	}

	bothttpClient := clients.NewBotClient(config.BotConfig.ScrapperBaseURL)

	bot, err := application.NewBot(config.BotConfig.TgToken, bothttpClient)
	if err != nil {
		fmt.Printf("Error creating bot: %v\n", err)
		return
	}

	go func() {
		bot.Start()
	}()

	serv := server.InitServer(config.BotConfig.Addr,
		server.InitBotRouting(),
		config.BotConfig.ReadTimeout,
		config.BotConfig.WriteTimeout)

	if err := serv.ListenAndServe(); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
