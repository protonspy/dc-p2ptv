package cli

import (
	"errors"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/protonspy/dc-p2ptv/server/internal/config"
)

const testSecret = "a shared secret"

// noSecret is the environment of a node whose operator forgot the secret.
func noSecret(string) string { return "" }

// withSecret is the environment of a node configured to serve.
func withSecret(name string) string {
	if name == SecretEnv {
		return testSecret
	}
	return ""
}

// stopAtOnce stands in for the signal a supervisor would send.
func stopAtOnce() {}

func TestTurnRejectsAnUnknownFlag(t *testing.T) {
	var stdout, stderr strings.Builder

	if got := runTURN(&stdout, &stderr, []string{"-nonsense"}, withSecret, stopAtOnce); got != Misused {
		t.Errorf("runTURN() = %d, want %d", got, Misused)
	}
}

func TestTurnRefusesToServeWithoutASecret(t *testing.T) {
	var stdout, stderr strings.Builder

	if got := runTURN(&stdout, &stderr, []string{"-relay-address", "127.0.0.1"}, noSecret, stopAtOnce); got != Misused {
		t.Errorf("runTURN() = %d, want %d", got, Misused)
	}
	if !strings.Contains(stderr.String(), SecretEnv) {
		t.Errorf("stderr = %q, want it to name %s", stderr.String(), SecretEnv)
	}
}

func TestTurnReportsAServiceItCannotStart(t *testing.T) {
	var stdout, stderr strings.Builder

	if got := runTURN(&stdout, &stderr, nil, withSecret, stopAtOnce); got != Failed {
		t.Errorf("runTURN() = %d, want %d", got, Failed)
	}
	if !strings.Contains(stderr.String(), "relay address") {
		t.Errorf("stderr = %q, want it to name the cause", stderr.String())
	}
}

func TestTurnServesAndReportsWhatFellBack(t *testing.T) {
	var stdout, stderr strings.Builder
	args := []string{"-listen", "127.0.0.1:0", "-relay-address", "127.0.0.1"}

	if got := runTURN(&stdout, &stderr, args, withSecret, stopAtOnce); got != OK {
		t.Fatalf("runTURN() = %d, want %d: %s", got, OK, stderr.String())
	}
	if !strings.Contains(stdout.String(), "federation=open own=0") {
		t.Errorf("stdout = %q, want it to report the posture it came up in", stdout.String())
	}
	if !strings.Contains(stdout.String(), "turn 127.0.0.1:") {
		t.Errorf("stdout = %q, want it to name where it listens", stdout.String())
	}
	if !strings.Contains(stdout.String(), "issued=0 relayed=0 fraction=0.000") {
		t.Errorf("stdout = %q, want it to report the share that fell back", stdout.String())
	}
	if stderr.String() != "" {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}
}

func TestTurnReportsAWriteFailure(t *testing.T) {
	var stderr strings.Builder
	broken := failingWriter{err: errors.New("device full")}
	args := []string{"-listen", "127.0.0.1:0", "-relay-address", "127.0.0.1"}

	if got := runTURN(broken, &stderr, args, withSecret, stopAtOnce); got != Failed {
		t.Errorf("runTURN() = %d, want %d", got, Failed)
	}
	if !strings.Contains(stderr.String(), "device full") {
		t.Errorf("stderr = %q, want it to carry the cause", stderr.String())
	}
}

// Run has to reach the subcommand; the missing secret is what stops it before it
// binds anything or waits for a signal.
func TestRunDispatchesTheTurnSubcommand(t *testing.T) {
	var stdout, stderr strings.Builder
	t.Setenv(SecretEnv, "")

	if got := Run(&stdout, &stderr, []string{"turn"}); got != Misused {
		t.Errorf("Run() = %d, want %d", got, Misused)
	}
	if !strings.Contains(stderr.String(), SecretEnv) {
		t.Errorf("stderr = %q, want it to name %s", stderr.String(), SecretEnv)
	}
}

func TestWaitForSignalEndsOnAnInterrupt(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a process cannot send itself an interrupt on Windows")
	}
	// Held for the whole test so that an interrupt arriving before the wait is
	// registered is delivered here instead of ending the test process.
	guard := make(chan os.Signal, 1)
	signal.Notify(guard, os.Interrupt)
	defer signal.Stop(guard)

	done := make(chan struct{})
	go func() {
		waitForSignal()
		close(done)
	}()

	self, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("finding this process: %v", err)
	}
	deadline := time.After(10 * time.Second)
	for {
		if err := self.Signal(os.Interrupt); err != nil {
			t.Fatalf("interrupting this process: %v", err)
		}
		select {
		case <-done:
			return
		case <-deadline:
			t.Fatal("waitForSignal did not return on an interrupt")
		case <-time.After(20 * time.Millisecond):
		}
	}
}

// The switch that closes the network is meant to be flipped without a release, so
// a value nobody can read has to stop the node rather than be guessed at.
func TestTurnRefusesAFederationValueItCannotRead(t *testing.T) {
	var stdout, stderr strings.Builder
	environment := func(name string) string {
		if name == config.FederationEnv {
			return "clsoed"
		}
		return withSecret(name)
	}

	got := runTURN(&stdout, &stderr, []string{"-relay-address", "127.0.0.1"}, environment, stopAtOnce)
	if got != Misused {
		t.Errorf("runTURN() = %d, want %d", got, Misused)
	}
	if !strings.Contains(stderr.String(), config.FederationEnv) {
		t.Errorf("stderr = %q, want it to name %s", stderr.String(), config.FederationEnv)
	}
}

// A supervisor asking what the subcommand takes has not misused it.
func TestTurnReportsItsFlagsOnRequest(t *testing.T) {
	var stdout, stderr strings.Builder

	if got := runTURN(&stdout, &stderr, []string{"-h"}, withSecret, stopAtOnce); got != OK {
		t.Errorf("runTURN() = %d, want %d", got, OK)
	}
	if !strings.Contains(stderr.String(), "-relay-address") {
		t.Errorf("stderr = %q, want it to name the flags", stderr.String())
	}
}

func TestTurnServesWithPrivatePeersAllowed(t *testing.T) {
	var stdout, stderr strings.Builder
	args := []string{"-listen", "127.0.0.1:0", "-relay-address", "127.0.0.1", "-allow-private-peers"}

	if got := runTURN(&stdout, &stderr, args, withSecret, stopAtOnce); got != OK {
		t.Errorf("runTURN() = %d, want %d: %s", got, OK, stderr.String())
	}
}
