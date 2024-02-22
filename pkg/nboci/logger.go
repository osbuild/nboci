package nboci

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

var emptyAttr = slog.Attr{}

// InitLogger initializes slog logging package
func InitLogger(level slog.Level) {
	th := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return emptyAttr
			}
			if a.Key == slog.LevelKey {
				return emptyAttr
			}
			return a
		},
	})
	logger := slog.New(th)
	slog.SetDefault(logger)
}

// SlogWriter writes Go standard library logger to slog.
type SlogWriter struct {
	Logger  *slog.Logger
	Level   slog.Level
}

func (slw SlogWriter) Write(p []byte) (n int, err error) {
	slw.Logger.Log(context.Background(), slw.Level, strings.TrimSpace(string(p)))

	return len(p), nil
}
