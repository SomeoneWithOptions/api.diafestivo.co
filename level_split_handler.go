package main

import (
	"context"
	"log/slog"
)

type levelSplitHandler struct {
	stdoutHandler slog.Handler
	stderrHandler slog.Handler
	errorLevel    slog.Level
}

func newLevelSplitHandler(stdoutHandler slog.Handler, stderrHandler slog.Handler, errorLevel slog.Level) slog.Handler {
	return &levelSplitHandler{
		stdoutHandler: stdoutHandler,
		stderrHandler: stderrHandler,
		errorLevel:    errorLevel,
	}
}

func (h *levelSplitHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if level >= h.errorLevel {
		return h.stderrHandler.Enabled(ctx, level)
	}

	return h.stdoutHandler.Enabled(ctx, level)
}

func (h *levelSplitHandler) Handle(ctx context.Context, record slog.Record) error {
	if record.Level >= h.errorLevel {
		return h.stderrHandler.Handle(ctx, record)
	}

	return h.stdoutHandler.Handle(ctx, record)
}

func (h *levelSplitHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &levelSplitHandler{
		stdoutHandler: h.stdoutHandler.WithAttrs(attrs),
		stderrHandler: h.stderrHandler.WithAttrs(attrs),
		errorLevel:    h.errorLevel,
	}
}

func (h *levelSplitHandler) WithGroup(name string) slog.Handler {
	return &levelSplitHandler{
		stdoutHandler: h.stdoutHandler.WithGroup(name),
		stderrHandler: h.stderrHandler.WithGroup(name),
		errorLevel:    h.errorLevel,
	}
}
