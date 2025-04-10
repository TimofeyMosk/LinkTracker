package main

import (
	"context"
	"errors"
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
		slog.Error("Error reading config", "error", err)
		return
	}

	pkg.InitLogger(config.ScrapConfig.LogsPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userRepo, linkRepo, stateManager, err := InitRepositories(ctx, config.DBConfig)
	if err != nil {
		slog.Error("Error initializing repositories", "error", err)
		return
	}

	botHTTPClient, err := clients.NewBotHTTPClient(config.ScrapConfig.BotBaseURL, config.ScrapConfig.BotClientTimeout)
	if err != nil {
		slog.Error("Error creating bot client", "error", err)
		return
	}

	linkSourceHandlers := InitLinksSourceHandlers()
	linkChecker := linkchecker.NewLinkChecker(linkRepo, linkSourceHandlers,
		config.ScrapConfig.SizeLinksPage,
		config.ScrapConfig.CheckerLinksWorkers,
	)

	messageNotifier := notifier.NewHTTPNotifier(botHTTPClient)

	scrap := scrapper.NewScrapper(userRepo, linkRepo, stateManager,
		config.ScrapConfig.Interval,
		messageNotifier,
		linkChecker,
	)

	serv := server.InitServer(
		config.ScrapConfig.Address,
		server.InitScrapperRouting(scrap),
		config.ScrapConfig.ReadTimeout,
		config.ScrapConfig.WriteTimeout,
	)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		application.StopScrapperSignalReceiving(ctx, cancel, serv)
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		if err := scrap.Run(ctx); err != nil {
			slog.Error("Scrapper run failed", "error", err)
		}
	}()

	if err := serv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Server failed to start or finished with error", "error", err)
	} else {
		slog.Info("Server stopped gracefully")
	}

	wg.Wait()
}
