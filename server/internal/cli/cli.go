// Package cli turns process arguments into an exit status. It exists so the
// entrypoint holds a single statement and every branch that decides anything is
// reachable from a test.
package cli

import (
	"fmt"
	"io"

	"github.com/protonspy/dc-p2ptv/server/internal/version"
)

// Exit statuses. A node is run by a supervisor, so these are part of its
// contract rather than an implementation detail.
const (
	OK      = 0
	Failed  = 1
	Misused = 2
)

// Run writes what this build is to stdout and reports the status the process
// should exit with. Anything it cannot do is reported on stderr.
func Run(stdout, stderr io.Writer, args []string) int {
	if len(args) > 0 {
		// Nothing useful remains if reporting the misuse also fails.
		_, _ = fmt.Fprintf(stderr, "node: unexpected argument %q\n", args[0])
		return Misused
	}
	if _, err := fmt.Fprintln(stdout, version.Describe("node")); err != nil {
		_, _ = fmt.Fprintln(stderr, "node:", err)
		return Failed
	}
	return OK
}
