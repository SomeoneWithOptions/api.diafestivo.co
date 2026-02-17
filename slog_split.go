package main

import (
	"context"
	"log/slog"
)

type splitHandler struct {
	stdoutHandler slog.Handler
	stderrHandler slog.Handler
	errorLevel    slog.Level
}

func newSplitHandler(stdoutHandler slog.Handler, stderrHandler slog.Handler, errorLevel slog.Level) slog.Handler {
	return &splitHandler{
		stdoutHandler: stdoutHandler,
		stderrHandler: stderrHandler,
		errorLevel:    errorLevel,
	}
}

func (h *splitHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if level >= h.errorLevel {
		return h.stderrHandler.Enabled(ctx, level)
	}

	return h.stdoutHandler.Enabled(ctx, level)
}

func (h *splitHandler) Handle(ctx context.Context, record slog.Record) error {
	if record.Level >= h.errorLevel {
		return h.stderrHandler.Handle(ctx, record)
	}

	return h.stdoutHandler.Handle(ctx, record)
}

func (h *splitHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &splitHandler{
		stdoutHandler: h.stdoutHandler.WithAttrs(attrs),
		stderrHandler: h.stderrHandler.WithAttrs(attrs),
		errorLevel:    h.errorLevel,
	}
}

func (h *splitHandler) WithGroup(name string) slog.Handler {
	return &splitHandler{
		stdoutHandler: h.stdoutHandler.WithGroup(name),
		stderrHandler: h.stderrHandler.WithGroup(name),
		errorLevel:    h.errorLevel,
	}
}
