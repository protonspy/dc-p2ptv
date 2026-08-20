package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/protonspy/dc-p2ptv/server/internal/config"
	"github.com/protonspy/dc-p2ptv/server/internal/nat"
)

// SecretEnv names the environment variable the TURN shared secret is read from.
// It is not a flag: a flag would put the secret in the process table of every
// machine the node runs on.
const SecretEnv = "NODE_TURN_SECRET"

// runTURN stands the TURN service up and serves until the process is asked to
// stop. Everything the outside world provides — the environment and the wait for
// a signal — is a parameter, so every branch here is reachable from a test.
func runTURN(stdout, stderr io.Writer, args []string, getenv func(string) string, wait func()) int {
	flags := flag.NewFlagSet("node turn", flag.ContinueOnError)
	flags.SetOutput(stderr)
	listen := flags.String("listen", nat.DefaultListenAddress, "address to serve TURN on")
	realm := flags.String("realm", nat.DefaultRealm, "realm announced to clients")
	relay := flags.String("relay-address", "", "public IP clients are told to relay through")
	ttl := flags.Duration("credential-ttl", nat.DefaultCredentialTTL, "how long an issued credential stays valid")
	private := flags.Bool("allow-private-peers", false, "relay towards addresses off the public internet, for a demonstration on one machine")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return OK
		}
		return Misused
	}

	secret := getenv(SecretEnv)
	if secret == "" {
		_, _ = fmt.Fprintf(stderr, "node: %s is empty\n", SecretEnv)
		return Misused
	}

	federation, err := config.FederationFrom(getenv)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "node:", err)
		return Misused
	}

	service, err := nat.Start(nat.Config{
		Realm:             *realm,
		SharedSecret:      secret,
		ListenAddress:     *listen,
		RelayAddress:      *relay,
		CredentialTTL:     *ttl,
		AllowPrivatePeers: *private,
	})
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "node:", err)
		return Failed
	}

	// A node says what posture it is in on the way up, because the switch that
	// closes the network is meant to be checked without shipping a release.
	if _, err := fmt.Fprintf(stdout, "turn %s realm=%s relay=%s %s\n",
		service.Address(), *realm, *relay, federation); err != nil {
		_, _ = fmt.Fprintln(stderr, "node:", err)
		_ = service.Close()
		return Failed
	}

	wait()

	issued, relayed, _ := service.Usage().Counts()
	fraction := service.Usage().Fraction()
	if err := service.Close(); err != nil {
		_, _ = fmt.Fprintln(stderr, "node:", err)
		return Failed
	}
	// The share that fell back to TURN is the number this service exists to keep
	// small, so it is reported where the operator will see it: on the way out.
	_, _ = fmt.Fprintf(stdout, "turn stopped issued=%d relayed=%d fraction=%.3f\n",
		issued, relayed, fraction)
	return OK
}

// waitForSignal blocks until the process is asked to stop. A node is run by a
// supervisor, so both the interrupt of a terminal and the termination of a
// supervisor have to end it.
func waitForSignal() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
}
