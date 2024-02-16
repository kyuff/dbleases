package dbleases

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kyuff/dbleases/internal/lease"
	"github.com/kyuff/dbleases/internal/split"
)

type Logger interface {
	InfofContext(ctx context.Context, template string, args ...any)
	ErrorfContext(ctx context.Context, template string, args ...any)
}

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type Repository interface {
	InsertLease(ctx context.Context, clientID string, leaseName string, value int, ttl time.Duration, status lease.Status) error
	GetAndRefreshLeases(ctx context.Context, names []string, clientID string, ttl time.Duration) ([]lease.Info, error)
	SetLeaseStatus(ctx context.Context, clientID string, leaseName string, value int, status lease.Status) error
	DeleteLeases(ctx context.Context, clientID string) error
}

func NewClient(db DB, clientID string, options ...Option) (*Client, error) {
	o := defaultOptions()
	for _, opt := range options {
		opt(&o)
	}

	if err := o.validate(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), o.migrationTimeout)
	defer cancel()
	repo, err := o.repositoryFactory(ctx, db)
	if err != nil {
		return nil, err
	}

	return &Client{
		ID:                clientID,
		db:                db,
		opt:               o,
		repo:              repo,
		heartbeatStopChan: make(chan struct{}),
		leases:            make(map[string]*Lease),
	}, nil
}

type Client struct {
	ID   string
	db   DB
	opt  Options
	repo Repository

	closed atomic.Bool

	heartbeatStopChan chan struct{}
	heartbeatStart    sync.Once

	leaseMux   sync.RWMutex
	leases     map[string]*Lease
	leaseNames []string
}

func (c *Client) Lease(name string, size int) *Lease {
	c.leaseMux.Lock()
	defer c.leaseMux.Unlock()

	if l, ok := c.leases[name]; ok {
		return l
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.opt.heartbeatTimeout)
	defer cancel()

	c.leases[name] = newLease(c, name, int(size))
	c.leaseNames = append(c.leaseNames, name)

	err := c.registerLease(ctx, lease.Request{
		ClientID:  c.ID,
		LeaseName: name,
		Value:     c.leases[name].ringValue[0],
		Status:    lease.Pending,
	})
	if err != nil {
		c.opt.logger.ErrorfContext(context.Background(), "Failed to register lease %q for client %s: %s", name, c.ID, err)
	}

	c.heartbeatStart.Do(c.startHeartbeat)
	return c.leases[name]
}

func (c *Client) Close() error {
	if c.closed.Load() {
		return nil
	}
	c.heartbeatStopChan <- struct{}{}
	<-c.heartbeatStopChan
	c.closed.Store(true)
	return nil
}

func (c *Client) startHeartbeat() {
	pump := func() {
		ctx, cancel := context.WithTimeout(context.Background(), c.opt.heartbeatTimeout)
		defer cancel()
		err := c.heartbeat(ctx)
		if err != nil {
			c.opt.logger.ErrorfContext(ctx, "[dbleases] Heartbeat failed: %s", err)
		}
	}

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), c.opt.heartbeatTimeout)
		defer cancel()
		err := c.cleanup(ctx)
		if err != nil {
			c.opt.logger.ErrorfContext(ctx, "[dbleases] Failed to close client %s: %s", c.ID, err)
		}
		c.heartbeatStopChan <- struct{}{}
	}

	go func() {
		ticker := time.NewTicker(c.opt.heartbeat)
		defer ticker.Stop()

		pump()

		defer cleanup()

		for {
			select {
			case <-c.heartbeatStopChan:
				return
			case <-ticker.C:
				pump()
			}
		}
	}()
}

func (c *Client) heartbeat(ctx context.Context) error {
	c.leaseMux.RLock()
	defer c.leaseMux.RUnlock()

	allLeases, err := c.repo.GetAndRefreshLeases(ctx, c.leaseNames, c.ID, c.opt.ttl)
	if err != nil {
		return err
	}
	leasesByName := split.By(lease.Ring(allLeases), func(lease lease.Info) string {
		return lease.Name
	})

	for name, leases := range leasesByName {
		l, ok := c.leases[name]
		if !ok {
			c.opt.logger.ErrorfContext(ctx, "[dbleases] Unexpected lease: %s", name)
			continue
		}

		report := leases.Analyze(c.ID, l.size)

		l.setValues(ctx, report.Values)
		c.approveLeases(ctx, report.Approvals)
		c.registerLeaseRequests(ctx, report.Balance)
	}

	return nil
}

func (c *Client) registerLeaseRequests(ctx context.Context, request *lease.Request) {
	//c.opt.logger.InfofContext(ctx, "[dbleases] Balancing %#v", request)
	if request == nil {
		return
	}

	err := c.registerLease(ctx, *request)
	if err != nil {
		c.opt.logger.ErrorfContext(ctx, "[dbleases] Failed to register lease: %s", err)
	}
}

func (c *Client) registerLease(ctx context.Context, r lease.Request) error {
	return c.repo.InsertLease(ctx, r.ClientID, r.LeaseName, r.Value, c.opt.ttl, r.Status)
}

func (c *Client) approveLeases(ctx context.Context, approvals []lease.Info) {
	if len(approvals) == 0 {
		return
	}

	for _, a := range approvals {
		err := c.repo.SetLeaseStatus(ctx, a.ClientID, a.Name, a.Value, lease.Leased)
		if err != nil {
			c.opt.logger.ErrorfContext(ctx, "[dbleases] Failed approving pending lease %q for client %s: %s", a.Name, a.ClientID, err)
		}
	}
}

func (c *Client) cleanup(ctx context.Context) error {
	c.leaseMux.Lock()
	defer c.leaseMux.Unlock()

	for _, l := range c.leases {
		l.setValues(ctx, nil)
	}

	return c.repo.DeleteLeases(ctx, c.ID)
}
