package application

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

func StopScrapperSignalReceiving(ctx context.Context, cancel context.CancelFunc, server *http.Server) {
	<-SignalWarden(syscall.SIGINT, syscall.SIGTERM)

	ctxServerShutdown, cancelServerShutdown := context.WithTimeout(ctx, 10*time.Second)
	defer cancelServerShutdown()

	err := server.Shutdown(ctxServerShutdown)
	if err != nil {
		slog.Error(err.Error())
	}

	cancel()
}

func StopBotSignalReceiving(ctx context.Context, cancel context.CancelFunc, server *http.Server) {
	<-SignalWarden(syscall.SIGINT, syscall.SIGTERM)

	ctxServerShutdown, cancelServerShutdown := context.WithTimeout(ctx, 10*time.Second)
	defer cancelServerShutdown()

	err := server.Shutdown(ctxServerShutdown)
	if err != nil {
		slog.Error(err.Error())
	}

	cancel()
}
