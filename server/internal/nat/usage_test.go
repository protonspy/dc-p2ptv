package nat

import (
	"sync"
	"testing"
)

func TestUsageCountsIssuedAndRelayedSessions(t *testing.T) {
	var usage Usage

	usage.Issued()
	usage.Issued()
	usage.Issued()
	usage.Allocated("a")
	usage.Allocated("b")

	issued, relayed, live := usage.Counts()
	if issued != 3 || relayed != 2 || live != 2 {
		t.Errorf("Counts() = %d, %d, %d, want 3, 2, 2", issued, relayed, live)
	}
}

func TestUsageCountsASessionOnceHoweverManyAllocationsItOpens(t *testing.T) {
	var usage Usage

	usage.Issued()
	usage.Allocated("a")
	usage.Allocated("a")
	usage.Allocated("a")

	issued, relayed, live := usage.Counts()
	if issued != 1 || relayed != 1 || live != 1 {
		t.Errorf("Counts() = %d, %d, %d, want 1, 1, 1", issued, relayed, live)
	}
}

func TestUsageReleasesASessionWithItsLastAllocation(t *testing.T) {
	var usage Usage

	usage.Issued()
	usage.Allocated("a")
	usage.Allocated("a")
	usage.Released("a")

	if _, _, live := usage.Counts(); live != 1 {
		t.Errorf("live = %d after one of two allocations went away, want 1", live)
	}

	usage.Released("a")

	issued, relayed, live := usage.Counts()
	if issued != 1 || relayed != 1 || live != 0 {
		t.Errorf("Counts() = %d, %d, %d, want 1, 1, 0", issued, relayed, live)
	}
}

func TestUsageIgnoresAReleaseNobodyAllocated(t *testing.T) {
	var usage Usage

	usage.Released("never seen")

	if _, _, live := usage.Counts(); live != 0 {
		t.Errorf("live = %d, want 0", live)
	}
}

func TestUsageFractionIsTheShareThatFellBack(t *testing.T) {
	var usage Usage

	for range 4 {
		usage.Issued()
	}
	usage.Allocated("a")

	if got := usage.Fraction(); got != 0.25 {
		t.Errorf("Fraction() = %v, want 0.25", got)
	}
}

func TestUsageFractionIsZeroWithNothingIssued(t *testing.T) {
	var usage Usage

	if got := usage.Fraction(); got != 0 {
		t.Errorf("Fraction() = %v, want 0", got)
	}
}

// A credential issued by another node's control plane relays here without ever
// having been counted as issued here, so the numerator can outrun the denominator.
func TestUsageFractionIsCappedAtOne(t *testing.T) {
	var usage Usage

	usage.Issued()
	usage.Allocated("a")
	usage.Allocated("b")

	if got := usage.Fraction(); got != 1 {
		t.Errorf("Fraction() = %v, want 1", got)
	}
}

// The TURN server reports allocations from its own goroutines while the control
// plane issues credentials from whichever request asked for them.
func TestUsageIsSafeForConcurrentUse(t *testing.T) {
	var usage Usage
	var wg sync.WaitGroup

	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			usage.Issued()
			session := string(rune('a' + i%7))
			usage.Allocated(session)
			usage.Released(session)
		}()
	}
	wg.Wait()

	issued, _, live := usage.Counts()
	if issued != 50 {
		t.Errorf("issued = %d, want 50", issued)
	}
	if live != 0 {
		t.Errorf("live = %d, want 0", live)
	}
}
