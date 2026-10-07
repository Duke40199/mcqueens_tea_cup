package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// ANSI escape codes for colored console output.
const (
	ansiReset  = "\033[0m"
	ansiDim    = "\033[2m"
	ansiBold   = "\033[1m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
)

// prettyHandler is a slog.Handler that writes compact, colored, human-readable
// lines for local development, e.g.:
//
//	14:22:07.141 INFO  starting OBMeta sync  trace_id=808de797 source=main.go:107
//
// It colors the level, dims the timestamp and attributes, shortens the "source"
// path to file:line, and omits empty-string attributes (e.g. an unset trace_id)
// to cut noise. Color is auto-disabled when the output is not a terminal.
type prettyHandler struct {
	mu    *sync.Mutex
	out   io.Writer
	level slog.Leveler
	color bool
	attrs []slog.Attr
}

func newPrettyHandler(out io.Writer, level slog.Leveler) *prettyHandler {
	color := false
	if f, ok := out.(*os.File); ok {
		if info, err := f.Stat(); err == nil && info.Mode()&os.ModeCharDevice != 0 {
			color = true
		}
	}
	return &prettyHandler{mu: &sync.Mutex{}, out: out, level: level, color: color}
}

func (h *prettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	var sb strings.Builder

	// Timestamp (dim).
	h.paint(&sb, ansiDim, r.Time.Format("15:04:05.000"))
	sb.WriteByte(' ')

	// Level (colored, bold, fixed width).
	color, text := h.levelStyle(r.Level)
	h.paint(&sb, color+ansiBold, fmt.Sprintf("%-5s", text))
	sb.WriteByte(' ')

	// Message.
	sb.WriteString(r.Message)

	// Attributes (dim, key=value), skipping empty strings.
	var parts []string
	appendAttr := func(a slog.Attr) {
		v := a.Value.Any()
		// Skip empty strings (e.g. unset trace_id/user_id) to cut noise.
		if s, ok := v.(string); ok && s == "" {
			return
		}
		parts = append(parts, fmt.Sprintf("%s=%v", a.Key, v))
	}
	for _, a := range h.attrs {
		appendAttr(a)
	}
	r.Attrs(func(a slog.Attr) bool {
		appendAttr(a)
		return true
	})
	if len(parts) > 0 {
		sb.WriteByte(' ')
		h.paint(&sb, ansiDim, strings.Join(parts, " "))
	}
	sb.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.out, sb.String())
	return err
}

func (h *prettyHandler) paint(sb *strings.Builder, color, s string) {
	if h.color {
		sb.WriteString(color)
		sb.WriteString(s)
		sb.WriteString(ansiReset)
	} else {
		sb.WriteString(s)
	}
}

func (h *prettyHandler) levelStyle(l slog.Level) (string, string) {
	switch {
	case l >= slog.LevelError:
		return ansiRed, "ERROR"
	case l >= slog.LevelWarn:
		return ansiYellow, "WARN"
	case l >= slog.LevelInfo:
		return ansiGreen, "INFO"
	default:
		return ansiCyan, "DEBUG"
	}
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nh := *h
	nh.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &nh
}

// WithGroup is a no-op: this project does not use attribute groups.
func (h *prettyHandler) WithGroup(_ string) slog.Handler { return h }
