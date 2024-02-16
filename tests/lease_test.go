package tests

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/kyuff/dbleases"
	"github.com/kyuff/dbleases/internal/assert"
)

func TestLeases(t *testing.T) {
	t.Parallel()
	var (
		db          = Connect(t)
		newClientID = func() string {
			return fmt.Sprintf("client-%06d", rand.Intn(100000))
		}
		newLeaseName = func() string {
			return fmt.Sprintf("lease-%06d", rand.Intn(100000))
		}
		newClient = func(t *testing.T, clientID string, overrideOptions ...dbleases.Option) *dbleases.Client {

			var testOptions = append([]dbleases.Option{
				dbleases.WithHeartbeat(time.Millisecond * 250),
				dbleases.WithTTL(time.Millisecond * 1000),
			}, overrideOptions...)
			client, err := dbleases.NewClient(db, clientID, testOptions...)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			t.Cleanup(func() {
				assert.NoError(t, client.Close())
			})

			return client
		}
		sequence = func(t *testing.T, plays ...func(t *testing.T)) {
			t.Helper()
			for _, fn := range plays {
				fn(t)
			}
		}
		parallel = func(plays ...func(t *testing.T)) func(t *testing.T) {
			run := func(t *testing.T, wg *sync.WaitGroup, fn func(t *testing.T)) {
				t.Helper()
				fn(t)
				wg.Done()
			}
			return func(t *testing.T) {
				t.Helper()
				var wg sync.WaitGroup
				wg.Add(len(plays))
				for i := 0; i < len(plays); i++ {
					go run(t, &wg, plays[i])
				}
				wg.Wait()
			}
		}
		assertValues = func(client *dbleases.Client, name string, size int, values []int) func(t *testing.T) {
			return func(t *testing.T) {
				t.Helper()
				lease := client.Lease(name, size)
				if !assert.EqualSliceWithin(t, time.Second*5, values, lease.Values) {
					t.Logf("client %q", client.ID)
				}
			}
		}
		startLease = func(client *dbleases.Client, name string, size int) func(t *testing.T) {
			return func(t *testing.T) {
				t.Helper()
				client.Lease(name, size)
			}
		}
		closeClient = func(client *dbleases.Client) func(t *testing.T) {
			return func(t *testing.T) {
				t.Helper()
				assert.NoError(t, client.Close())
			}
		}
	)

	t.Run("should create the client", func(t *testing.T) {
		t.Parallel()
		// arrange
		var (
			clientID  = newClientID()
			leaseName = newLeaseName()
			client    = newClient(t, clientID)
		)

		// act
		_ = client.Lease(leaseName, 100)
	})

	t.Run("should lease a full range", func(t *testing.T) {
		t.Parallel()
		// arrange
		var (
			clientID  = newClientID()
			leaseName = newLeaseName()
			client    = newClient(t, clientID)
		)

		// act
		lease := client.Lease(leaseName, 3)

		// assert
		assert.EqualSliceWithin(t, time.Second*2, []int{0, 1, 2}, lease.Values)
	})

	t.Run("should split lease", func(t *testing.T) {
		t.Parallel()
		// arrange
		var (
			leaseName = newLeaseName()
			clients   = []*dbleases.Client{
				newClient(t, "bl hash 3/15"),
				newClient(t, "cl hash 5/15"),
				newClient(t, "ay hash 9/15"),
				newClient(t, "aj hash 14/15"),
			}
			leases  []*dbleases.Lease
			expects = [][]int{
				{3, 4},
				{5, 6, 7, 8},
				{9, 10, 11, 12, 13},
				{0, 1, 2, 14},
			}
			wg sync.WaitGroup
		)

		wg.Add(len(clients))

		// act
		for _, client := range clients {
			leases = append(leases, client.Lease(leaseName, 15))
		}

		// assert
		for i := 0; i < len(leases); i++ {
			go func(lease *dbleases.Lease, expect []int) {
				assert.EqualSliceWithin(t, time.Second*2, expect, lease.Values)
				wg.Done()
			}(leases[i], expects[i])
		}

		wg.Wait()
	})

	t.Run("should ignore a hash clash", func(t *testing.T) {
		t.Parallel()
		// arrange
		var (
			leaseName    = newLeaseName()
			firstClient  = newClient(t, "ak hash 3/5")
			secondClient = newClient(t, "as hash 3/5")
		)

		firstLease := firstClient.Lease(leaseName, 5)

		// act
		secondLease := secondClient.Lease(leaseName, 5)

		// assert
		assert.EqualSliceWithin(t, time.Second*2, []int{0, 1, 2, 3, 4}, firstLease.Values)
		assert.EqualSliceWithin(t, time.Second*2, []int{}, secondLease.Values)
	})

	t.Run("should scale up to 3 clients and back down", func(t *testing.T) {
		t.Parallel()
		// arrange
		var (
			leaseName = newLeaseName()
			leaseSize = 20
			clientA   = newClient(t, "bd hash 5/20")
			clientB   = newClient(t, "ap hash 10/20")
			clientC   = newClient(t, "an hash 15/20")
		)

		// act - assert
		sequence(t,
			// scale up
			startLease(clientA, leaseName, leaseSize),
			assertValues(clientA, leaseName, leaseSize, fromTo(0, 19)),
			startLease(clientB, leaseName, leaseSize),
			parallel(
				assertValues(clientA, leaseName, leaseSize, fromTo(5, 9)),
				assertValues(clientB, leaseName, leaseSize, join(fromTo(0, 4), fromTo(10, 19))),
			),
			startLease(clientC, leaseName, leaseSize),
			parallel(
				assertValues(clientA, leaseName, leaseSize, fromTo(5, 9)),
				assertValues(clientB, leaseName, leaseSize, fromTo(10, 14)),
				assertValues(clientC, leaseName, leaseSize, join(fromTo(0, 4), fromTo(15, 19))),
			),
			// scale down
			closeClient(clientB),
			parallel(
				assertValues(clientA, leaseName, leaseSize, fromTo(5, 14)),
				assertValues(clientB, leaseName, leaseSize, nil),
				assertValues(clientC, leaseName, leaseSize, join(fromTo(0, 4), fromTo(15, 19))),
			),
			closeClient(clientA),
			parallel(
				assertValues(clientA, leaseName, leaseSize, nil),
				assertValues(clientB, leaseName, leaseSize, nil),
				assertValues(clientC, leaseName, leaseSize, fromTo(0, 19)),
			),
			closeClient(clientC),
			parallel(
				assertValues(clientA, leaseName, leaseSize, nil),
				assertValues(clientB, leaseName, leaseSize, nil),
				assertValues(clientC, leaseName, leaseSize, nil),
			),
		)
	})

	t.Run("should balance the values", func(t *testing.T) {
		t.Parallel()
		// arrange
		var (
			leaseName = newLeaseName()
			leaseSize = 100
			clientA   = newClient(t, "bj hash 15/100")
			clientB   = newClient(t, "ik hash 15/100")
			clientC   = newClient(t, "is hash 15/100")
		)

		// act - assert
		sequence(t,
			// scale up
			startLease(clientA, leaseName, leaseSize),
			assertValues(clientA, leaseName, leaseSize, fromTo(0, 99)),
			startLease(clientB, leaseName, leaseSize),
			parallel(
				assertValues(clientA, leaseName, leaseSize, fromTo(15, 64)),
				assertValues(clientB, leaseName, leaseSize, join(fromTo(0, 14), fromTo(65, 99))),
			),
			startLease(clientC, leaseName, leaseSize),
			parallel(
				assertValues(clientA, leaseName, leaseSize, fromTo(10, 43)),
				assertValues(clientB, leaseName, leaseSize, join(fromTo(0, 6), fromTo(65, 90), fromTo(99, 99))),
				assertValues(clientC, leaseName, leaseSize, join(fromTo(7, 9), fromTo(44, 64), fromTo(91, 98))),
			),
			// scale down
			closeClient(clientC),
			parallel(
				assertValues(clientA, leaseName, leaseSize, fromTo(10, 59)),
				assertValues(clientB, leaseName, leaseSize, join(fromTo(0, 9), fromTo(60, 99))),
			),
			parallel(
				closeClient(clientA),
				closeClient(clientB),
			),
		)
	})
}

func fromTo(from, to int) []int {
	var items []int
	for i := from; i <= to; i++ {
		items = append(items, i)
	}
	return items
}

func join(lists ...[]int) []int {
	var r []int
	for _, list := range lists {
		r = append(r, list...)
	}
	return r
}
