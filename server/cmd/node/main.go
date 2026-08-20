// Command node runs one node of the network: control plane, forwarding, or both.
package main

import (
	"os"

	"github.com/protonspy/dc-p2ptv/server/internal/cli"
)

func main() { os.Exit(cli.Run(os.Stdout, os.Stderr, os.Args[1:])) }
