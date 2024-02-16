package dbleases

import (
	"context"
	"sync"

	"github.com/kyuff/dbleases/internal/hash"
)

func newLease(client *Client, name string, size int) *Lease {
	return &Lease{
		client:    client,
		name:      name,
		size:      size,
		ringValue: []int{hash.Mod(client.ID, size)},
	}
}

type Lease struct {
	client *Client
	name   string
	size   int

	ringValue []int

	mu     sync.RWMutex
	values []int
}

func (m *Lease) Values() []int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.values
}

func (m *Lease) setValues(ctx context.Context, values []int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !equal(m.values, values) {
		m.client.opt.logger.InfofContext(ctx, "[dbleases] Lease %q for client %q set to %q", m.name, m.client.ID, presentIntegers(values))
	}

	m.values = values
}

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
