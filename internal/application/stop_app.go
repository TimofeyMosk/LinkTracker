package application

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	err = server.Shutdown(context.Background())
	if err != nil {
		slog.Error(err.Error())
	}
}

func StopBotSignalReceiving(bot *Bot, server *http.Server) {
	<-SignalWarden(syscall.SIGINT, syscall.SIGTERM)

	bot.Stop()

	err := server.Shutdown(context.Background())
	if err != nil {
		slog.Error(err.Error())
	}
}
