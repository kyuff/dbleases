# Database Leases for Go

Library to support shared leases across multiple processes using a database.

## Use case

In a distributed system, it is common to have workload spread across multiple processes. One example is multiple pods in
Kubernetes. For some of these types of workloads, it can be crucial to avoid workers competing for the same work.
One strategy to avoid that is to use a database lock, that prevents two workers taking the same work. That is not
always practical and it comes with a cost to the database.

This library offers another solution to cases where the work can be partitioned. With the library you can get a list of
partition values that can be used to only handle the work assigned to your partitions.

## FAQ

### Which database and drivers is supported?

Postgres is supported. The code is tested with [pgx](https://github.com/jackc/pgx), but is assumed to work
with [lib/pg](https://github.com/lib/pq) as well.

### What happens if a client is removed forcefully?

The assigned values will continue to be leased until the TTL runs out. At that time another client will take over.

### How long should my TTL be?

It depends on the type of workload you have and the load on your database. In simple terms, it's a trade-off between
having partitions locked too long vs load on your database.

A high TTL will allow you a similar high heartbeat rate, which in returns results in a lower load on your database.
A low TTL will allow for clients that disappear ungracefully to release their lock faster, at the cost of database load.

### What is the default TTL/Heartbeat?

TTL is 6 seconds and heartbeat is 5 seconds. It can be seen [here](options.go).

### Is there a problem with clock sync between multiple machines and the TTL?

No. The client code newer uses the internal clock in the host performing the work. All time operations is done in the
database, which removes the problem of clock drift.

### Is work evenly distributed between clients?

Yes, there is a simple mechanism that distributes the partitions between clients.

### What is a good partition size?

It depend son how many clients you expect to distribute the partitions over. A high number allows for less risk of
conflicts between clients initially being assigned the same partitions. It frees the library from having to solve the
conflict. A good starting point is around 100 partitions for 10 or fewer clients.

### Can I call the dbleases.Lease.Values() method in a busy loop?

Yes! It is concurrency safe and is (near) free to call.

The method will return a different list of integers when the lease/client is repartitioned.

## Example

````go
package example

import (
	"database/sql"
	"errors"
	"os"
	"time"

	"github.com/kyuff/dbleases"
)

func Setup(db *sql.DB) error {
	var (
		clientID      = os.Getenv("HOSTNAME")
		leaseName     = "workload-type"
		partitionSize = 128
	)
	client, err := dbleases.NewClient(db, clientID,
		dbleases.WithHeartbeat(time.Second*4),
		dbleases.WithTTL(time.Second*5),
	)
	if err != nil {
		return err
	}

	defer func() {
		_ = client.Close()
	}()

	lease := client.Lease(leaseName, partitionSize)

	for {
		err = performWorkItem(db, lease.Values())
		if err != nil {
			return err
		}
	}
}

func performWorkItem(db *sql.DB, partitions []int) error {
	// Add worker code here
	return errors.New("not implemented")
}

````
