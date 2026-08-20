package nat

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/pion/logging"
	"github.com/pion/turn/v4"
)

func TestConfigRejectsAnEmptySecret(t *testing.T) {
	if _, _, err := (Config{RelayAddress: "127.0.0.1"}).withDefaults(); !errors.Is(err, ErrNoSecret) {
		t.Errorf("withDefaults() error = %v, want %v", err, ErrNoSecret)
	}
}

func TestConfigRejectsANegativeCredentialLifetime(t *testing.T) {
	cfg := Config{SharedSecret: testSecret, RelayAddress: "127.0.0.1", CredentialTTL: -time.Second}

	if _, _, err := cfg.withDefaults(); !errors.Is(err, ErrTTLNotPositive) {
		t.Errorf("withDefaults() error = %v, want %v", err, ErrTTLNotPositive)
	}
}

func TestConfigRejectsAnEmptyRelayAddress(t *testing.T) {
	if _, _, err := (Config{SharedSecret: testSecret}).withDefaults(); !errors.Is(err, ErrNoRelayAddress) {
		t.Errorf("withDefaults() error = %v, want %v", err, ErrNoRelayAddress)
	}
}

// A hostname here is the mistake that looks like a working server nobody reaches:
// the relay address travels to the client as an IP and is never resolved.
func TestConfigRejectsARelayAddressThatIsNotAnIP(t *testing.T) {
	cfg := Config{SharedSecret: testSecret, RelayAddress: "turn.example.com"}

	if _, _, err := cfg.withDefaults(); !errors.Is(err, ErrRelayAddressInvalid) {
		t.Errorf("withDefaults() error = %v, want %v", err, ErrRelayAddressInvalid)
	}
}

func TestConfigFillsTheRealmAndTheListenAddress(t *testing.T) {
	cfg, relay, err := (Config{SharedSecret: testSecret, RelayAddress: "203.0.113.7"}).withDefaults()
	if err != nil {
		t.Fatalf("withDefaults() error = %v", err)
	}

	if cfg.Realm != DefaultRealm {
		t.Errorf("Realm = %q, want %q", cfg.Realm, DefaultRealm)
	}
	if cfg.ListenAddress != DefaultListenAddress {
		t.Errorf("ListenAddress = %q, want %q", cfg.ListenAddress, DefaultListenAddress)
	}
	if !relay.Equal(net.ParseIP("203.0.113.7")) {
		t.Errorf("relay = %v, want 203.0.113.7", relay)
	}
}

func TestStartReportsWhatItCannotBind(t *testing.T) {
	cfg := Config{SharedSecret: testSecret, RelayAddress: "127.0.0.1", ListenAddress: "127.0.0.1:70000"}

	service, err := Start(cfg)
	if err == nil {
		_ = service.Close()
		t.Fatalf("Start() error = nil, want one")
	}
}

func TestStartReportsAConfigurationItCannotUse(t *testing.T) {
	service, err := Start(Config{SharedSecret: testSecret})
	if err == nil {
		_ = service.Close()
		t.Fatalf("Start() error = nil, want one")
	}
}

// The whole point of the service, end to end: a client holding a credential this
// node issued relays through it, and the session is counted as one that fell back.
func TestServiceRelaysForACredentialItIssued(t *testing.T) {
	service := mustService(t)
	credential, err := service.Credentials().Issue("session")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	client, done := dial(t, service, credential.Username, credential.Password)
	defer done()

	relay, err := client.Allocate()
	if err != nil {
		t.Fatalf("Allocate() error = %v", err)
	}

	issued, relayed, live := service.Usage().Counts()
	if issued != 1 || relayed != 1 || live != 1 {
		t.Errorf("Counts() = %d, %d, %d, want 1, 1, 1", issued, relayed, live)
	}
	if got := service.Usage().Fraction(); got != 1 {
		t.Errorf("Fraction() = %v, want 1", got)
	}

	if err := relay.Close(); err != nil {
		t.Fatalf("closing the relay: %v", err)
	}
	waitForLive(t, service, 0)
}

func TestServiceRefusesACredentialItDidNotIssue(t *testing.T) {
	service := mustService(t)
	credential, err := service.Credentials().Issue("session")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	client, done := dial(t, service, credential.Username, "not the password")
	defer done()

	if relay, err := client.Allocate(); err == nil {
		_ = relay.Close()
		t.Fatalf("Allocate() error = nil, want one")
	}

	if _, relayed, _ := service.Usage().Counts(); relayed != 0 {
		t.Errorf("relayed = %d, want 0", relayed)
	}
}

func TestServiceAddressNamesThePortItGot(t *testing.T) {
	service := mustService(t)

	if _, port, err := net.SplitHostPort(service.Address().String()); err != nil || port == "0" {
		t.Errorf("Address() = %v, want a bound port", service.Address())
	}
}

func TestCloseIsReportedWhenItFails(t *testing.T) {
	service := mustService(t)

	if err := service.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := service.Close(); err == nil {
		t.Errorf("closing twice reported nothing, want the second to report")
	}
}

func mustService(t *testing.T) *Service {
	t.Helper()
	service, err := Start(Config{
		SharedSecret:  testSecret,
		ListenAddress: "127.0.0.1:0",
		RelayAddress:  "127.0.0.1",
		CredentialTTL: time.Minute,
	})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() { _ = service.Close() })
	return service
}

func dial(t *testing.T, service *Service, username, password string) (*turn.Client, func()) {
	t.Helper()
	conn, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listening for the client: %v", err)
	}
	client, err := turn.NewClient(&turn.ClientConfig{
		TURNServerAddr: service.Address().String(),
		Username:       username,
		Password:       password,
		Realm:          DefaultRealm,
		Conn:           conn,
		LoggerFactory:  logging.NewDefaultLoggerFactory(),
	})
	if err != nil {
		_ = conn.Close()
		t.Fatalf("NewClient() error = %v", err)
	}
	if err := client.Listen(); err != nil {
		client.Close()
		_ = conn.Close()
		t.Fatalf("Listen() error = %v", err)
	}
	return client, func() {
		client.Close()
		_ = conn.Close()
	}
}

// waitForLive polls, because the allocation goes away on the server's own
// goroutine after the client says so.
func waitForLive(t *testing.T, service *Service, want uint64) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, _, live := service.Usage().Counts(); live == want {
			return
		}
		if time.Now().After(deadline) {
			_, _, live := service.Usage().Counts()
			t.Fatalf("live = %d after waiting, want %d", live, want)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestPermitPeerRefusesWhatThePublicInternetDoesNotRoute(t *testing.T) {
	permit := permitPeer(false)

	for _, peer := range []string{"127.0.0.1", "10.0.0.4", "192.168.1.9", "169.254.1.1", "224.0.0.1", "0.0.0.0", "::1", "fe80::1"} {
		if permit(nil, net.ParseIP(peer)) {
			t.Errorf("permitPeer(false) admitted %s, want it refused", peer)
		}
	}
	if !permit(nil, net.ParseIP("203.0.113.7")) {
		t.Errorf("permitPeer(false) refused a public address, want it admitted")
	}
	if permit(nil, nil) {
		t.Errorf("permitPeer(false) admitted an address it could not read, want it refused")
	}
}

// A demonstration on one machine is the case that legitimately relays to loopback.
func TestPermitPeerAdmitsPrivateAddressesWhenAsked(t *testing.T) {
	permit := permitPeer(true)

	if !permit(nil, net.ParseIP("127.0.0.1")) {
		t.Errorf("permitPeer(true) refused loopback, want it admitted")
	}
}

// End to end: the node forwards for whoever holds a credential, so a credential
// must not be a way into the network the node itself can reach.
func TestServiceRefusesToRelayIntoItsOwnNetwork(t *testing.T) {
	service := mustService(t)
	credential, err := service.Credentials().Issue("session")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	client, done := dial(t, service, credential.Username, credential.Password)
	defer done()

	relay, err := client.Allocate()
	if err != nil {
		t.Fatalf("Allocate() error = %v", err)
	}
	defer func() { _ = relay.Close() }()

	peer := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9}
	if err := client.CreatePermission(peer); err == nil {
		t.Errorf("CreatePermission(loopback) error = nil, want the service to refuse")
	}
}
