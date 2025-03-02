package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/server"
)

func main() {
	config, err := application.ReadYAMLConfig("config.yaml")
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		return
	}

	application.InitLogger(config.LogsPath)

	wg := sync.WaitGroup{}
	rep := repository.NewRepository()
	scrapClient := clients.NewScrapperClient(config.ScrapConfig.BotBaseURL)
	scrapper := application.NewScrapper(rep, config.ScrapConfig.Interval, scrapClient)
	ser := server.InitServer(
		config.ScrapConfig.Addr,
		server.InitScrapperRouting(scrapper),
		config.ScrapConfig.ReadTimeout,
		config.ScrapConfig.WriteTimeout)

	go application.StopSignalReceiving(scrapper, ser)

	wg.Add(1)

	go func() {
		if err := scrapper.Run(); err != nil {
			slog.Error(err.Error())
		}

		wg.Done()
	}()

	if err := ser.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Println("server failed to start or finished with error", err)
	} else {
		fmt.Println("application stopped")
	}

	wg.Wait()
}
