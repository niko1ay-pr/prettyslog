package prettyslog

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"sync"
)

const (
	timeFormat = "[15:04:05.000]"
	reset      = "\033[0m"

	black = iota + 30
	red
	green
	yellow
	blue
	magenta
	cyan
	lightGray
	darkGray = iota + 90
	lightRed
	lightGreen
	lightYellow
	lightBlue
	lightMagenta
	lightCyan
	white
)

type Handler struct {
	h slog.Handler
	b *bytes.Buffer
	m *sync.Mutex
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	h.h.Enabled(ctx, level)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{h: h.h.WithAttrs(attrs), b: h.b, m: h.m}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{h: h.h.WithGroup(name), b: h.b, m: h.m}
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = colorize(darkGray, level)
	case slog.LevelError:
		level = colorize(lightRed, level)
	case slog.LevelInfo:
		level = colorize(cyan, level)
	case slog.LevelWarn:
		level = colorize(lightYellow, level)
	}

	fmt.Println(colorize(lightGray, r.Time.Format(timeFormat)), level, colorize(white, r.Message))

	return nil
}

func colorize(colorCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", colorCode, v, reset)
}
