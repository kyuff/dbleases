package lease

import (
	"fmt"
	"math"
	"sort"
	"time"
)

const (
	balanceThreshold = 5
)

type Status string

func (s Status) String() string {
	return string(s)
}

const (
	Pending Status = "PENDING"
	Leased  Status = "LEASED"
)

type Info struct {
	Name     string
	ClientID string
	TTL      time.Time
	Status   Status
	Value    int
}

func (i Info) String() string {
	return fmt.Sprintf("[%s] %q %s %q %d", i.Name, i.ClientID, i.TTL.Format("15.04.05.999999999"), i.Status, i.Value)
}

type Request struct {
	ClientID  string
	LeaseName string
	Value     int
	Status    Status
}

type Report struct {
	Values    []int
	Approvals []Info
	Balance   *Request
}

type Ring []Info

func (ring Ring) nextLease(i int) Info {
	if len(ring) == 0 {
		return Info{}
	}
	if len(ring) == 1 {
		return ring[0]
	}

	if len(ring) <= i+1 {
		return ring[0]
	}

	return ring[i+1]
}

// Analyze a Ring to Report the values for a clientID in a given size.
//
// This method follows a set of rules that all clients are expected to follow.
// The implicit rules allows balancing a Ring and avoids conflicts where multiple
// nodes competes for a a given value.
//
// The rules are:
// - A Ring is a sequence of numbers starting at 0 and monotonously increasing
// - A Ring continuous from the highest number to 0, giving an endless loop
// - A Client has a lease on it's value all values up to but not including the next in the Ring.
// - If a Ring only consists of Pending clients, the lowest value client must approve itself.
// - It is the responsibility of the previous Leased Client to approve the next Pending Client
// - A Client can lease values by adding a value in a Pending state
// - A solo Client has a lease on the entire Ring
// - If there is a need to Balance it must be done by the client with the least value values or the lowest value
func (ring Ring) Analyze(clientID string, size int) Report {
	var (
		report    Report
		clients   = make(map[string]Client)
		leaseName string
	)
	sort.Slice(ring, func(i, j int) bool { return ring[i].Value < ring[j].Value })

	if len(ring) == 0 {
		return report
	}

	for i, lease := range ring {
		var (
			next      = ring.nextLease(i)
			approvals []Info
			values    []int
		)

		switch {
		case lease.Status == Leased && next.Status == Pending:
			approvals = append(approvals, next)
		case lease.Status == Pending && next.Status == Pending:
			if lease.Value <= next.Value && clientID == lease.ClientID {
				approvals = append(approvals, lease)
			}
		}

		leaseName = lease.Name

		if lease.Status == Leased {
			if next.Value > lease.Value {
				values = append(values, fromTo(lease.Value, next.Value)...)
			} else {
				values = append(values, fromTo(0, next.Value)...)
				values = append(values, fromTo(lease.Value, size)...)
			}
		}

		if lease.ClientID == clientID {
			report.Approvals = append(report.Approvals, approvals...)
			report.Values = append(report.Values, values...)
		}

		if _, ok := clients[lease.ClientID]; !ok {
			clients[lease.ClientID] = Client{
				ID: lease.ClientID,
			}
		}

		clients[lease.ClientID] = Client{ID: lease.ClientID,
			Values: append(clients[lease.ClientID].Values, values...),
		}
	}

	if len(clients) < size {
		report.Balance = analyzeBalance(leaseName, clientID, clients)
	}

	sort.Ints(report.Values)

	return report
}

type Client struct {
	ID     string
	Values []int
}

func analyzeBalance(leaseName string, clientID string, clients map[string]Client) *Request {
	var (
		maxClientID string
		minClientID string
		maxSize     int
		minSize     = math.MaxInt
	)
	for clientID, client := range clients {
		size := len(client.Values)
		if size > maxSize {
			maxSize = size
			maxClientID = clientID
		}
		if size < minSize {
			minSize = size
			minClientID = clientID
		}
	}

	sizeDiff := maxSize - minSize
	_, clientInRing := clients[clientID]
	clientIsBehind := minClientID == clientID && sizeDiff > balanceThreshold
	systemIsStarted := maxSize > 0

	if systemIsStarted && (!clientInRing || clientIsBehind) {
		adjust := sizeDiff / 2
		if !clientInRing {
			adjust = 1
		}

		return &Request{
			ClientID:  clientID,
			LeaseName: leaseName,
			Value:     clients[maxClientID].Values[maxSize-adjust],
			Status:    Pending,
		}
	}

	return nil
}

func fromTo(from, to int) []int {
	var items []int
	for i := from; i < to; i++ {
		items = append(items, i)
	}
	return items
}
