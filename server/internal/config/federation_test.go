package config

import (
	"errors"
	"slices"
	"testing"
)

// environment stands in for the process the node happens to be running in.
func environment(values map[string]string) func(string) string {
	return func(name string) string { return values[name] }
}

func TestFederationDefaultsToOpen(t *testing.T) {
	federation, err := FederationFrom(environment(nil))
	if err != nil {
		t.Fatalf("FederationFrom() error = %v", err)
	}

	if !federation.IsOpen() {
		t.Errorf("IsOpen() = false, want true with nothing configured")
	}
}

func TestFederationClosesOnTheKey(t *testing.T) {
	federation, err := FederationFrom(environment(map[string]string{FederationEnv: " CLOSED "}))
	if err != nil {
		t.Fatalf("FederationFrom() error = %v", err)
	}

	if federation.IsOpen() {
		t.Errorf("IsOpen() = true, want false with %s closed", FederationEnv)
	}
}

func TestFederationStaysOpenOnTheExplicitValue(t *testing.T) {
	federation, err := FederationFrom(environment(map[string]string{FederationEnv: Open}))
	if err != nil {
		t.Fatalf("FederationFrom() error = %v", err)
	}

	if !federation.IsOpen() {
		t.Errorf("IsOpen() = false, want true")
	}
}

// A misspelled value read as open is a network that federates while the operator
// believes it does not, so it is refused rather than guessed.
func TestFederationRefusesAValueItDoesNotKnow(t *testing.T) {
	_, err := FederationFrom(environment(map[string]string{FederationEnv: "clsoed"}))

	if !errors.Is(err, ErrFederationUnknown) {
		t.Errorf("FederationFrom() error = %v, want %v", err, ErrFederationUnknown)
	}
}

func TestOwnNodesAreReadInAStableOrder(t *testing.T) {
	federation, err := FederationFrom(environment(map[string]string{
		OwnNodesEnv: " node-b , node-a ,, node-b ",
	}))
	if err != nil {
		t.Fatalf("FederationFrom() error = %v", err)
	}

	if want := []string{"node-a", "node-b"}; !slices.Equal(federation.Own(), want) {
		t.Errorf("Own() = %v, want %v", federation.Own(), want)
	}
}

func TestAnOpenNetworkAdmitsAnyNodeThatIdentifiesItself(t *testing.T) {
	federation, err := FederationFrom(environment(nil))
	if err != nil {
		t.Fatalf("FederationFrom() error = %v", err)
	}

	if !federation.Admits("some-other-node") {
		t.Errorf("Admits() = false, want true on an open network")
	}
}

func TestAClosedNetworkAdmitsOnlyItsOwnNodes(t *testing.T) {
	federation, err := FederationFrom(environment(map[string]string{
		FederationEnv: Closed,
		OwnNodesEnv:   "node-a,node-b",
	}))
	if err != nil {
		t.Fatalf("FederationFrom() error = %v", err)
	}

	if !federation.Admits("node-a") {
		t.Errorf("Admits(node-a) = false, want true")
	}
	if federation.Admits("node-c") {
		t.Errorf("Admits(node-c) = true, want false on a closed network")
	}
}

// Unidentified is not the same as unrestricted, on either posture.
func TestNoPostureAdmitsANodeWithoutAnIdentity(t *testing.T) {
	for _, posture := range []string{Open, Closed} {
		federation, err := FederationFrom(environment(map[string]string{FederationEnv: posture}))
		if err != nil {
			t.Fatalf("FederationFrom() error = %v", err)
		}
		if federation.Admits("   ") {
			t.Errorf("Admits(blank) = true on a %s network, want false", posture)
		}
	}
}

func TestFederationReportsItsPosture(t *testing.T) {
	federation, err := FederationFrom(environment(map[string]string{
		FederationEnv: Closed,
		OwnNodesEnv:   "node-a,node-b",
	}))
	if err != nil {
		t.Fatalf("FederationFrom() error = %v", err)
	}

	if want := "federation=closed own=2"; federation.String() != want {
		t.Errorf("String() = %q, want %q", federation.String(), want)
	}
}

func TestAnOpenFederationReportsItself(t *testing.T) {
	federation, err := FederationFrom(environment(nil))
	if err != nil {
		t.Fatalf("FederationFrom() error = %v", err)
	}

	if want := "federation=open own=0"; federation.String() != want {
		t.Errorf("String() = %q, want %q", federation.String(), want)
	}
}
