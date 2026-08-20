// Command node runs one node of the network: control plane, forwarding, or both.
package main

import (
	"fmt"

	"github.com/protonspy/dc-p2ptv/server/internal/version"
)

func main() {
	fmt.Println(version.Describe("node"))
}
