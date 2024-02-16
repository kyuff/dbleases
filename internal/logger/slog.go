package logger

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"
)

func NewSlog(logger *slog.Logger) *Slog {
	return &Slog{logger: logger}
}

type Slog struct {
	logger *slog.Logger
}

func (l *Slog) InfofContext(ctx context.Context, template string, args ...any) {
	if !l.logger.Enabled(ctx, slog.LevelInfo) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, InfofContext]
	r := slog.NewRecord(time.Now(), slog.LevelInfo, fmt.Sprintf(template, args...), pcs[0])
	_ = l.logger.Handler().Handle(ctx, r)
}

func (l *Slog) ErrorfContext(ctx context.Context, template string, args ...any) {
	if !l.logger.Enabled(ctx, slog.LevelError) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, ErrorfContext]
	r := slog.NewRecord(time.Now(), slog.LevelInfo, fmt.Sprintf(template, args...), pcs[0])
	_ = l.logger.Handler().Handle(ctx, r)
}
