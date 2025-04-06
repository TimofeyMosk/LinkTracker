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

func StopScrapperSignalReceiving(cancel context.CancelFunc, server *http.Server) {
	<-SignalWarden(syscall.SIGINT, syscall.SIGTERM)
	cancel()

	err := server.Shutdown(context.TODO())
	if err != nil {
		slog.Error(err.Error())
	}
}

func StopBotSignalReceiving(cancel context.CancelFunc, server *http.Server) {
	<-SignalWarden(syscall.SIGINT, syscall.SIGTERM)

	cancel()

	err := server.Shutdown(context.TODO())
	if err != nil {
		slog.Error(err.Error())
	}
}
