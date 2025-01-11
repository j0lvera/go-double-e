package testutils

import (
	"github.com/charmbracelet/log"
	"log/slog"
	"os"
)

func SetupTestLogger() {
	handler := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		Level:           log.DebugLevel,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
	handler.SetLevel(log.DebugLevel) // force debug level
}
