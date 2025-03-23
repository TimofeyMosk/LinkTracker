package application

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"LinkTracker/internal/infrastructure/clients"
)

func SignalWarden(signals ...os.Signal) chan struct{} {
	result := make(chan struct{})

	go func() {
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, signals...)

		<-signalCh
		close(result)
	}()

	return result
}

func StopScrapperSignalReceiving(scrapper *Scrapper, server *http.Server) {
	<-SignalWarden(syscall.SIGINT, syscall.SIGTERM)

	err := scrapper.Stop()
	if err != nil {
		slog.Error(err.Error())
	}

	err = server.Shutdown(context.TODO())
	if err != nil {
		slog.Error(err.Error())
	}

	slog.Info("soft shutdown was a success")
}

func StopBotSignalReceiving(tgBotAPI *clients.TelegramHTTPClient, server *http.Server) {
	<-SignalWarden(syscall.SIGINT, syscall.SIGTERM)

	tgBotAPI.Stop()

	err := server.Shutdown(context.TODO())
	if err != nil {
		slog.Error(err.Error())
	}
	slog.Info("soft shutdown was a success")
}
