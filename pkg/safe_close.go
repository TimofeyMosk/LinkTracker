package pkg

import (
	"io"
	"log/slog"
)

func SafeClose(c io.Closer) {
	if err := c.Close(); err != nil {
		slog.Error("could not close resource", "error", err.Error())
	}
}
