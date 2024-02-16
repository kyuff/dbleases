package dbleases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	logger2 "github.com/kyuff/dbleases/internal/logger"
	"github.com/kyuff/dbleases/internal/schemas/postgres"
)

type Option func(o *Options)

type Options struct {
	ttl               time.Duration
	heartbeat         time.Duration
	migrationTimeout  time.Duration
	heartbeatTimeout  time.Duration
	repositoryFactory func(ctx context.Context, db DB) (Repository, error)
	logger            Logger
}

func (opt Options) validate() error {
	if opt.heartbeat >= opt.ttl {
		return fmt.Errorf("[dbleases] heartbeat too slow compared to ttl: %s >= %s", opt.heartbeat, opt.ttl)
	}

	if opt.ttl < time.Second {
		return fmt.Errorf("[dbleases] ttl must be higher than 1 second: %s", opt.ttl)
	}

	return nil
}

func defaultOptions() Options {
	var opts = []Option{
		WithTTL(time.Second * 6),
		WithPostgres("public", "db_leases"),
		WithHeartbeat(time.Second * 5),
		WithMigrationTimeout(time.Second * 5),
		withHeartbeatTimeout(time.Second),
		WithSlog(slog.Default()),
	}
	var o Options
	for _, opt := range opts {
		opt(&o)
	}

	return o
}

func WithPostgres(schema, tablePrefix string) Option {
	return func(o *Options) {
		o.repositoryFactory = func(ctx context.Context, db DB) (Repository, error) {
			return postgres.New(ctx, db, schema, tablePrefix)
		}
	}
}

func WithHeartbeat(heartbeat time.Duration) Option {
	return func(o *Options) {
		o.heartbeat = heartbeat
	}
}

func WithMigrationTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.migrationTimeout = timeout
	}
}

func withHeartbeatTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.heartbeatTimeout = timeout
	}
}

func WithTTL(ttl time.Duration) Option {
	return func(o *Options) {
		o.ttl = ttl
	}
}

func WithSlog(l *slog.Logger) Option {
	return func(o *Options) {
		o.logger = logger2.NewSlog(l)
	}
}

func WithLoggingDisabled() Option {
	return func(o *Options) {
		o.logger = logger2.NewNoop()
	}
}
