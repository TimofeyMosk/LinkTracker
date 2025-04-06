package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"LinkTracker/internal/application"
	"LinkTracker/internal/application/scrapper"
	"LinkTracker/internal/application/scrapper/linkchecker"
	"LinkTracker/internal/application/scrapper/notifier"
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

	pkg.InitLogger(config.ScrapConfig.LogsPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userRepo, linkRepo, stateManager, err := InitRepositories(ctx, config.DBConfig)
	if err != nil {
		fmt.Printf("Error initializing repository: %v\n", err)
		return
	}

	botHTTPClient, err := clients.NewBotHTTPClient(config.ScrapConfig.BotBaseURL, config.ScrapConfig.BotClientTimeout)
	if err != nil {
		fmt.Printf("Error creating scrapper client: %v\n", err)
		return
	}

	messageNotifier := notifier.NewHTTPNotifier(botHTTPClient)

	linkSourceHandlers := InitLinksSourceHandlers()
	linkChecker := linkchecker.NewLinkChecker(linkRepo, linkSourceHandlers)

	scrap := scrapper.NewScrapper(userRepo, linkRepo, stateManager, config.ScrapConfig.Interval, messageNotifier, linkChecker)
	serv := server.InitServer(
		config.ScrapConfig.Address,
		server.InitScrapperRouting(scrap),
		config.ScrapConfig.ReadTimeout,
		config.ScrapConfig.WriteTimeout)

	go application.StopScrapperSignalReceiving(cancel, serv)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		if err := scrap.Run(ctx); err != nil {
			slog.Error(err.Error())
		}

		wg.Done()
	}()

	if err := serv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server failed to start or finished with error", "error", err)
	} else {
		slog.Info("server stopped")
	}

	wg.Wait()
}
