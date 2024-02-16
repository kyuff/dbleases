package logger

import (
	"context"
)

func NewNoop() Noop {
	return Noop{}
}

type Noop struct {
}

func (Noop) InfofContext(ctx context.Context, template string, args ...any) {}

func (Noop) ErrorfContext(ctx context.Context, template string, args ...any) {}
