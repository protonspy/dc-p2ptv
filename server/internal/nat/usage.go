// Package nat carries the network's NAT traversal: the TURN service a node runs
// as a safety net for the sessions ICE cannot connect directly, and the measure
// of how many sessions end up falling back to it.
//
// TURN is cost, never a data path. A rising fraction means the media the network
// exists to send directly is being paid for twice, so the measure is part of the
// service rather than something bolted on afterwards.
package nat

import "sync"

// Usage counts how many sessions were handed TURN credentials and how many of
// them actually relayed through the service. It is safe for concurrent use: the
// TURN server reports allocations from its own goroutines while the control
// plane issues credentials from the request that asked for them.
//
// A session that opens several allocations — one per peer connection — counts
// once, which is what makes the ratio a fraction of sessions rather than of
// allocations.
type Usage struct {
	mu      sync.Mutex
	issued  uint64
	relayed uint64
	live    map[string]int
}

// Issued records that a session was handed credentials. It is the denominator:
// every session that could have used the service, whether or not it did. A
// session that renews its credentials is counted again, because the renewal is
// another window in which it could have relayed.
func (u *Usage) Issued() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.issued++
}

// Allocated records that a session opened an allocation. Only the first
// allocation of a session moves the numerator; the rest are counted as live so
// that the session is released when the last of them goes away.
func (u *Usage) Allocated(session string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.live == nil {
		u.live = make(map[string]int)
	}
	if u.live[session] == 0 {
		u.relayed++
	}
	u.live[session]++
}

// Released records that one of a session's allocations went away. The session
// leaves the live set once its last allocation is gone, which is what keeps the
// set bounded by what the service is carrying rather than by what it has ever
// carried. A release nobody allocated is ignored.
func (u *Usage) Released(session string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.live[session] == 0 {
		return
	}
	u.live[session]--
	if u.live[session] == 0 {
		delete(u.live, session)
	}
}

// Counts reports the credentials issued, the sessions that relayed, and the
// sessions relaying right now, read as one consistent snapshot.
func (u *Usage) Counts() (issued, relayed, live uint64) {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.issued, u.relayed, uint64(len(u.live))
}

// Fraction is the share of sessions that fell back to TURN, between 0 and 1.
// With nothing issued there is no share to report, and it is 0 rather than a
// division by zero. Credentials issued by another node's control plane can make
// the numerator outrun the denominator, so the result is capped at 1.
func (u *Usage) Fraction() float64 {
	issued, relayed, _ := u.Counts()
	if issued == 0 {
		return 0
	}
	if relayed > issued {
		return 1
	}
	return float64(relayed) / float64(issued)
}
