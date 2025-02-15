package prettyslog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
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
	return h.h.Enabled(ctx, level)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{h: h.h.WithAttrs(attrs), b: h.b, m: h.m}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{h: h.h.WithGroup(name), b: h.b, m: h.m}
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	op := "handler.Handle"
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

	attr, err := h.extractAttrs(ctx, r)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	bytes, err := json.MarshalIndent(attr, "", "  ")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	fmt.Println(colorize(lightGray, r.Time.Format(timeFormat)),
		level,
		colorize(white, r.Message),
		colorize(darkGray, string(bytes)),
	)

	return nil
}

func NewHandler(opts *slog.HandlerOptions) *Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	b := &bytes.Buffer{}
	return &Handler{
		b: b,
		h: slog.NewJSONHandler(b, &slog.HandlerOptions{
			Level:       opts.Level,
			AddSource:   opts.AddSource,
			ReplaceAttr: supressDefaults(opts.ReplaceAttr),
		}),
		m: &sync.Mutex{},
	}
}

func colorize(colorCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, reset)
}

func (h *Handler) extractAttrs(ctx context.Context, r slog.Record) (map[string]any, error) {
	op := "hanler.extractAttrs"
	h.m.Lock()

	defer func() {
		h.b.Reset()
		h.m.Unlock()
	}()

	err := h.h.Handle(ctx, r)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var attrs map[string]any
	err = json.Unmarshal(h.b.Bytes(), &attrs)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return attrs, nil
}

func supressDefaults(next func([]string, slog.Attr) slog.Attr) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, attr slog.Attr) slog.Attr {
		if attr.Key == slog.TimeKey ||
			attr.Key == slog.LevelKey ||
			attr.Key == slog.MessageKey {
			return slog.Attr{}
		}
		if next == nil {
			return attr
		}

		return next(groups, attr)

	}
}
