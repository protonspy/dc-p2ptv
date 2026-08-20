// Package config reads what a node decides at runtime from its environment.
// Runtime rather than build time is the point: a switch that needs a release to
// flip is a switch nobody can reach on the night it is needed.
package config

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// The keys a node reads. FederationEnv is the one that closes the network to the
// operator's own nodes; OwnNodesEnv says which nodes those are, because until the
// node registry exists there is nowhere else for that to be written down.
const (
	FederationEnv = "NODE_FEDERATION"
	OwnNodesEnv   = "NODE_OWN_NODES"
)

// The values FederationEnv takes. Open is the default: the network federates,
// and closing it is the deliberate act.
const (
	Open   = "open"
	Closed = "closed"
)

// ErrFederationUnknown is reported for a value that is neither open nor closed.
// A misspelled value is never read as open: the failure mode of guessing here is
// a network that federates when the operator believed it did not.
var ErrFederationUnknown = errors.New("config: " + FederationEnv + " is neither " + Open + " nor " + Closed)

// Federation is the node's posture towards the rest of the network: whether it
// works with any node that identifies itself, or only with the operator's own.
type Federation struct {
	open bool
	own  map[string]struct{}
}

// FederationFrom reads the posture from the environment. The environment is a
// parameter so that a node can be configured from something other than the
// process it happens to be running in, and so that every branch here is
// reachable from a test.
func FederationFrom(getenv func(string) string) (Federation, error) {
	own := ownNodes(getenv(OwnNodesEnv))
	switch strings.ToLower(strings.TrimSpace(getenv(FederationEnv))) {
	case "", Open:
		return Federation{open: true, own: own}, nil
	case Closed:
		return Federation{open: false, own: own}, nil
	default:
		return Federation{}, fmt.Errorf("%w: %q", ErrFederationUnknown, getenv(FederationEnv))
	}
}

// ownNodes reads the comma-separated list of node identities the operator calls
// its own. Blanks are dropped rather than admitted as a node with no name.
func ownNodes(value string) map[string]struct{} {
	own := make(map[string]struct{})
	for _, node := range strings.Split(value, ",") {
		if node = strings.TrimSpace(node); node != "" {
			own[node] = struct{}{}
		}
	}
	return own
}

// IsOpen reports whether the node works with any node that identifies itself.
func (f Federation) IsOpen() bool { return f.open }

// Own names the nodes the operator calls its own, in a stable order so that what
// a node reports about itself does not change between reads.
func (f Federation) Own() []string {
	own := make([]string, 0, len(f.own))
	for node := range f.own {
		own = append(own, node)
	}
	sort.Strings(own)
	return own
}

// Admits reports whether this node will work with the named one. A node with no
// identity is admitted by neither posture: unidentified is not the same as
// unrestricted.
func (f Federation) Admits(node string) bool {
	if strings.TrimSpace(node) == "" {
		return false
	}
	if f.open {
		return true
	}
	_, own := f.own[node]
	return own
}

// String is what a node reports about its own posture on the way up, which is
// where an operator checks that the switch actually took.
func (f Federation) String() string {
	posture := Closed
	if f.open {
		posture = Open
	}
	return fmt.Sprintf("federation=%s own=%d", posture, len(f.own))
}
