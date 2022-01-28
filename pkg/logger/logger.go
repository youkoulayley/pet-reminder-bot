package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Setup setups the logger.
func Setup(level, format string) error {
	l, err := zerolog.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("parse log level: %w", err)
	}

	zerolog.SetGlobalLevel(l)

	var w io.Writer

	switch format {
	case "json":
		w = os.Stderr
	case "console":
		w = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		}
	default:
		w = os.Stderr
	}

	log.Logger = zerolog.New(w).With().Caller().Timestamp().Logger()

	return nil
}
