package nat

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/pion/logging"
	"github.com/pion/turn/v4"
)

// DefaultRealm is the realm a node announces when the configuration is silent.
// It is part of the credential's derivation, so it has to match on both sides.
const DefaultRealm = "dc-p2ptv"

// DefaultListenAddress is where the service listens when the configuration is
// silent: the port IANA assigns to TURN, on every interface.
const DefaultListenAddress = ":3478"

// Errors a service reports rather than starting into a shape that cannot work.
var (
	ErrNoRelayAddress      = errors.New("nat: the relay address is empty")
	ErrRelayAddressInvalid = errors.New("nat: the relay address is not an IP address")
)

// Config is everything a node needs to stand its TURN service up. The relay
// address is the one clients are told to send to, which on a machine behind a
// router is not the address the socket binds to — getting it wrong is the
// failure that looks like a working server nobody can reach.
type Config struct {
	Realm         string
	SharedSecret  string
	ListenAddress string
	RelayAddress  string
	CredentialTTL time.Duration

	// AllowPrivatePeers lets the service relay towards addresses that are not on
	// the public internet. It is off by default: a TURN server that will carry
	// traffic to a loopback or private address is a way into the network the node
	// runs on, for anyone holding a credential. A demonstration on one machine is
	// the case that legitimately turns it on.
	AllowPrivatePeers bool
}

// withDefaults fills what the configuration left silent and reports what it
// cannot decide for the operator.
func (c Config) withDefaults() (Config, net.IP, error) {
	if strings.TrimSpace(c.SharedSecret) == "" {
		return Config{}, nil, ErrNoSecret
	}
	if c.CredentialTTL < 0 {
		return Config{}, nil, ErrTTLNotPositive
	}
	if strings.TrimSpace(c.Realm) == "" {
		c.Realm = DefaultRealm
	}
	if strings.TrimSpace(c.ListenAddress) == "" {
		c.ListenAddress = DefaultListenAddress
	}
	if strings.TrimSpace(c.RelayAddress) == "" {
		return Config{}, nil, ErrNoRelayAddress
	}
	relay := net.ParseIP(c.RelayAddress)
	if relay == nil {
		return Config{}, nil, fmt.Errorf("%w: %q", ErrRelayAddressInvalid, c.RelayAddress)
	}
	return c, relay, nil
}

// Service is a running TURN server together with the issuer of the credentials
// it accepts and the measure of how much of the network is falling back to it.
// It decides nothing about a room: it moves bytes for whoever holds a credential
// the shared secret verifies, and it cannot read what it moves.
type Service struct {
	server      *turn.Server
	conn        net.PacketConn
	credentials *Credentials
	usage       *Usage
}

// Start binds the socket and serves TURN on it. The caller closes the service;
// closing it closes the socket the server was given.
func Start(cfg Config) (*Service, error) {
	cfg, relay, err := cfg.withDefaults()
	if err != nil {
		return nil, err
	}

	usage := &Usage{}
	credentials, err := NewCredentials(cfg.SharedSecret, cfg.CredentialTTL, usage)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenPacket("udp", cfg.ListenAddress)
	if err != nil {
		return nil, fmt.Errorf("nat: listening on %s: %w", cfg.ListenAddress, err)
	}

	logger := logging.NewDefaultLoggerFactory()
	server, err := turn.NewServer(turn.ServerConfig{
		Realm:         cfg.Realm,
		LoggerFactory: logger,
		AuthHandler:   turn.LongTermTURNRESTAuthHandler(cfg.SharedSecret, logger.NewLogger("turn")),
		EventHandler:  usage.eventHandler(),
		PacketConnConfigs: []turn.PacketConnConfig{{
			PacketConn: conn,
			RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
				RelayAddress: relay,
				Address:      "0.0.0.0",
			},
			PermissionHandler: permitPeer(cfg.AllowPrivatePeers),
		}},
	})
	if err != nil {
		// The socket is ours until the server takes it, and the server did not.
		_ = conn.Close()
		return nil, fmt.Errorf("nat: starting the TURN service: %w", err)
	}

	return &Service{server: server, conn: conn, credentials: credentials, usage: usage}, nil
}

// eventHandler wires the server's allocation lifecycle to the measurement. Only
// the two events that say a session did or stopped relaying are handled; the
// rest are the server's business.
func (u *Usage) eventHandler() turn.EventHandler {
	return turn.EventHandler{
		OnAllocationCreated: func(_, _ net.Addr, _, username, _ string, _ net.Addr, _ int) {
			u.Allocated(sessionOf(username))
		},
		OnAllocationDeleted: func(_, _ net.Addr, _, username, _ string) {
			u.Released(sessionOf(username))
		},
	}
}

// Credentials is the issuer for this service. What it mints, this service accepts.
func (s *Service) Credentials() *Credentials { return s.credentials }

// Usage is the measure of how many sessions fell back to this service.
func (s *Service) Usage() *Usage { return s.usage }

// Address is where the service is actually listening, which is what a caller
// that asked for port 0 needs in order to tell anyone about it.
func (s *Service) Address() net.Addr { return s.conn.LocalAddr() }

// Close stops serving and releases the socket.
func (s *Service) Close() error {
	if err := s.server.Close(); err != nil {
		return fmt.Errorf("nat: stopping the TURN service: %w", err)
	}
	return nil
}

// permitPeer decides which peers the service will carry traffic to. The node
// forwards for whoever holds a valid credential, so without this the service is
// a path from anywhere into whatever the node's own network can reach.
func permitPeer(allowPrivate bool) turn.PermissionHandler {
	return func(_ net.Addr, peer net.IP) bool {
		if allowPrivate {
			return true
		}
		return isPublic(peer)
	}
}

// isPublic reports whether an address is one the public internet routes to.
func isPublic(ip net.IP) bool {
	switch {
	case ip == nil, ip.IsUnspecified(), ip.IsLoopback(), ip.IsPrivate():
		return false
	case ip.IsLinkLocalUnicast(), ip.IsLinkLocalMulticast(), ip.IsMulticast():
		return false
	case ip.IsInterfaceLocalMulticast():
		return false
	default:
		return true
	}
}
