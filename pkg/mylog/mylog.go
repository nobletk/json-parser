package mylog

import (
	"log/slog"

	prettylog "github.com/nobletk/json-parser/pkg/pretty-log"
)

type MyLog struct {
	Debug  bool
	Logger *slog.Logger
}

func CreateLogger(debug bool) *slog.Logger {
	var level slog.Level

	if debug {
		level = slog.LevelDebug
	} else {
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: false,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "nothing" {
				return slog.Attr{}
			}
			return a
		},
	}
	logger := slog.New(prettylog.NewHandler(opts))

	logger = logger.WithGroup("data")

	return logger
}
