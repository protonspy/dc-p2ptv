package nat

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/pion/logging"
	"github.com/pion/turn/v4"
)

const testSecret = "a shared secret"

func TestNewCredentialsRejectsAnEmptySecret(t *testing.T) {
	if _, err := NewCredentials("   ", time.Minute, nil); !errors.Is(err, ErrNoSecret) {
		t.Errorf("NewCredentials() error = %v, want %v", err, ErrNoSecret)
	}
}

func TestNewCredentialsRejectsANegativeLifetime(t *testing.T) {
	if _, err := NewCredentials(testSecret, -time.Second, nil); !errors.Is(err, ErrTTLNotPositive) {
		t.Errorf("NewCredentials() error = %v, want %v", err, ErrTTLNotPositive)
	}
}

func TestNewCredentialsDefaultsTheLifetime(t *testing.T) {
	credentials, err := NewCredentials(testSecret, 0, nil)
	if err != nil {
		t.Fatalf("NewCredentials() error = %v", err)
	}

	credential, err := credentials.Issue("session")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	want := time.Now().Add(DefaultCredentialTTL)
	if credential.ExpiresAt.Sub(want).Abs() > time.Minute {
		t.Errorf("ExpiresAt = %v, want about %v", credential.ExpiresAt, want)
	}
}

func TestIssueRejectsAnEmptySession(t *testing.T) {
	credentials := mustCredentials(t, nil)

	if _, err := credentials.Issue("  "); !errors.Is(err, ErrNoSession) {
		t.Errorf("Issue() error = %v, want %v", err, ErrNoSession)
	}
}

// The username is timestamp:session, so a session carrying a colon would be read
// back as a different session than the one the credential was minted for.
func TestIssueRejectsASessionWithAColon(t *testing.T) {
	credentials := mustCredentials(t, nil)

	if _, err := credentials.Issue("room:one"); !errors.Is(err, ErrSessionColon) {
		t.Errorf("Issue() error = %v, want %v", err, ErrSessionColon)
	}
}

func TestIssueMintsACredentialTheServiceAccepts(t *testing.T) {
	credentials := mustCredentials(t, nil)

	credential, err := credentials.Issue("session")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if !strings.HasSuffix(credential.Username, ":session") {
		t.Errorf("Username = %q, want it to end in the session", credential.Username)
	}

	authenticate := turn.LongTermTURNRESTAuthHandler(testSecret, logging.NewDefaultLoggerFactory().NewLogger("test"))
	key, ok := authenticate(credential.Username, DefaultRealm, nil)
	if !ok {
		t.Fatalf("the service rejected the credential it should accept")
	}
	if want := turn.GenerateAuthKey(credential.Username, DefaultRealm, credential.Password); string(key) != string(want) {
		t.Errorf("the key the service derived does not match the password issued")
	}
}

func TestIssueRejectsACredentialAfterItExpires(t *testing.T) {
	credentials := mustCredentials(t, nil)
	credentials.ttl = -time.Hour // already over, without waiting an hour for it

	credential, err := credentials.Issue("session")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	authenticate := turn.LongTermTURNRESTAuthHandler(testSecret, logging.NewDefaultLoggerFactory().NewLogger("test"))
	if _, ok := authenticate(credential.Username, DefaultRealm, nil); ok {
		t.Errorf("the service accepted a credential whose window had closed")
	}
}

func TestIssueCountsTheSessionAsOneThatCouldHaveRelayed(t *testing.T) {
	var usage Usage
	credentials := mustCredentials(t, &usage)

	if _, err := credentials.Issue("session"); err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if issued, _, _ := usage.Counts(); issued != 1 {
		t.Errorf("issued = %d, want 1", issued)
	}
}

func TestIssueWorksWithNobodyMeasuring(t *testing.T) {
	credentials := mustCredentials(t, nil)

	if _, err := credentials.Issue("session"); err != nil {
		t.Errorf("Issue() error = %v", err)
	}
}

func TestSessionOfDropsTheTimestamp(t *testing.T) {
	if got := sessionOf("1700000000:session"); got != "session" {
		t.Errorf("sessionOf() = %q, want %q", got, "session")
	}
}

// A username in no shape this node minted is still one session, and counting it
// as the empty session would merge every such client into one.
func TestSessionOfKeepsAUsernameItCannotSplit(t *testing.T) {
	if got := sessionOf("session"); got != "session" {
		t.Errorf("sessionOf() = %q, want %q", got, "session")
	}
}

func TestExpiryOfRejectsAUsernameWithoutATimestamp(t *testing.T) {
	if _, err := expiryOf("session"); err == nil {
		t.Errorf("expiryOf() error = nil, want one")
	}
}

func TestExpiryOfRejectsATimestampThatIsNotANumber(t *testing.T) {
	if _, err := expiryOf("soon:session"); err == nil {
		t.Errorf("expiryOf() error = nil, want one")
	}
}

func mustCredentials(t *testing.T, usage *Usage) *Credentials {
	t.Helper()
	credentials, err := NewCredentials(testSecret, time.Minute, usage)
	if err != nil {
		t.Fatalf("NewCredentials() error = %v", err)
	}
	return credentials
}
