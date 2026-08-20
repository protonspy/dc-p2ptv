package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/protonspy/dc-p2ptv/server/internal/version"
)

// failingWriter reports what a closed pipe or a full disk would.
type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestRunReportsTheBuild(t *testing.T) {
	var stdout, stderr strings.Builder

	if got := Run(&stdout, &stderr, nil); got != OK {
		t.Errorf("Run() = %d, want %d", got, OK)
	}
	if want := "node/" + version.Current + "\n"; stdout.String() != want {
		t.Errorf("stdout = %q, want %q", stdout.String(), want)
	}
	if stderr.String() != "" {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunRejectsAnArgument(t *testing.T) {
	var stdout, stderr strings.Builder

	if got := Run(&stdout, &stderr, []string{"--serve"}); got != Misused {
		t.Errorf("Run() = %d, want %d", got, Misused)
	}
	if stdout.String() != "" {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), `unexpected argument "--serve"`) {
		t.Errorf("stderr = %q, want it to name the argument", stderr.String())
	}
}

func TestRunReportsAWriteFailure(t *testing.T) {
	var stderr strings.Builder
	broken := failingWriter{err: errors.New("device full")}

	if got := Run(broken, &stderr, nil); got != Failed {
		t.Errorf("Run() = %d, want %d", got, Failed)
	}
	if !strings.Contains(stderr.String(), "device full") {
		t.Errorf("stderr = %q, want it to carry the cause", stderr.String())
	}
}
