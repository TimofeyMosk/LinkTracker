package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"LinkTracker/internal/application"
	"LinkTracker/internal/infrastructure/clients"
	"LinkTracker/internal/infrastructure/repository"
	"LinkTracker/internal/infrastructure/server"
	"LinkTracker/pkg"
)

func main() {
	config, err := application.ReadYAMLConfig("config.yaml")
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		return
	}

	pkg.InitLogger(config.ScrapConfig.LogsPath)

	wg := sync.WaitGroup{}
	rep := repository.NewRepository()

	botHTTPClient, err := clients.NewBotHTTPClient(config.ScrapConfig.BotBaseURL, config.ScrapConfig.BotClientTimeout)
	if err != nil {
		fmt.Printf("Error creating scrapper client: %v\n", err)
		return
	}

	scrapper := application.NewScrapper(rep, config.ScrapConfig.Interval, botHTTPClient)
	serv := server.InitServer(
		config.ScrapConfig.Address,
		server.InitScrapperRouting(scrapper),
		config.ScrapConfig.ReadTimeout,
		config.ScrapConfig.WriteTimeout)

	go application.StopScrapperSignalReceiving(scrapper, serv)

	wg.Add(1)

	go func() {
		if err := scrapper.Run(); err != nil {
			slog.Error(err.Error())
		}

		wg.Done()
	}()

	if err := serv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Println("server failed to start or finished with error", err)
		slog.Error("server failed to start or finished with error", "error", err)
	} else {
		fmt.Println("server stopped")
		slog.Info("server stopped")
	}

	wg.Wait()
}
