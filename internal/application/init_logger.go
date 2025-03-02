package application

import (
	"fmt"
	"log/slog"
	"os"
)

func InitLogger(logsPath string) {
	file, err := os.OpenFile(logsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		fmt.Println(err)
	}

	logger := slog.New(slog.NewTextHandler(file, nil))
	slog.SetDefault(logger)
}
