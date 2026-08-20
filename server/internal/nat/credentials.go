package nat

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pion/turn/v4"
)

// DefaultCredentialTTL is how long an issued credential stays valid when the
// configuration does not say. It is short because a credential is handed out per
// session and a session that outlives it asks for another, while a leaked one
// stops working on its own.
const DefaultCredentialTTL = time.Hour

// Errors an issuer reports rather than handing out a credential that cannot work.
var (
	ErrNoSecret       = errors.New("nat: the shared secret is empty")
	ErrTTLNotPositive = errors.New("nat: the credential lifetime is not positive")
	ErrNoSession      = errors.New("nat: the session is empty")
	ErrSessionColon   = errors.New("nat: the session contains a colon")
)

// Credential is what a client puts in its ICE server list. It carries no secret
// of the node's: the password is derived from the shared secret and expires on
// its own, so a client that keeps it learns nothing and gains nothing.
type Credential struct {
	Username  string
	Password  string
	ExpiresAt time.Time
}

// Credentials issues the ephemeral credentials the TURN service accepts. The
// service verifies them against the same shared secret, so no credential is ever
// stored and the control plane can issue them without talking to the data plane
// — which is the separation the network is built on.
type Credentials struct {
	secret string
	ttl    time.Duration
	usage  *Usage
}

// NewCredentials builds an issuer over a shared secret. A zero lifetime takes
// DefaultCredentialTTL; a negative one is a mistake and is reported. The usage
// may be nil when nobody is measuring.
func NewCredentials(secret string, ttl time.Duration, usage *Usage) (*Credentials, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, ErrNoSecret
	}
	if ttl == 0 {
		ttl = DefaultCredentialTTL
	}
	if ttl < 0 {
		return nil, ErrTTLNotPositive
	}
	return &Credentials{secret: secret, ttl: ttl, usage: usage}, nil
}

// Issue mints a credential for one session and counts the session as one that
// could have relayed. The username carries the expiry the service checks, so the
// deadline travels with the credential instead of being remembered anywhere.
func (c *Credentials) Issue(session string) (Credential, error) {
	if strings.TrimSpace(session) == "" {
		return Credential{}, ErrNoSession
	}
	if strings.Contains(session, ":") {
		return Credential{}, ErrSessionColon
	}
	username, password, err := turn.GenerateLongTermTURNRESTCredentials(c.secret, session, c.ttl)
	if err != nil {
		return Credential{}, fmt.Errorf("nat: minting a credential for %q: %w", session, err)
	}
	expiry, err := expiryOf(username)
	if err != nil {
		return Credential{}, err
	}
	if c.usage != nil {
		c.usage.Issued()
	}
	return Credential{Username: username, Password: password, ExpiresAt: expiry}, nil
}

// sessionOf recovers the session from a credential username. The TURN server
// reports allocations by the username that authenticated them, and the
// measurement is per session, so the timestamp in front of it is dropped. A
// username in no known shape is its own session rather than an empty one.
func sessionOf(username string) string {
	if _, session, found := strings.Cut(username, ":"); found {
		return session
	}
	return username
}

// expiryOf reads the deadline a credential username carries.
func expiryOf(username string) (time.Time, error) {
	stamp, _, found := strings.Cut(username, ":")
	if !found {
		return time.Time{}, fmt.Errorf("nat: credential username %q carries no expiry", username)
	}
	seconds, err := strconv.ParseInt(stamp, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("nat: credential username %q carries no expiry: %w", username, err)
	}
	return time.Unix(seconds, 0).UTC(), nil
}
