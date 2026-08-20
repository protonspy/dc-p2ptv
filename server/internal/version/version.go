// Package version carries the build identity a node reports to the rest of the network.
package version

import "strings"

// Current is the version this build reports. A release overwrites it at link time.
var Current = "0.0.0-dev"

// Describe renders the version together with the role the process runs as. The
// network's registry keys on this string, so an empty role is reported rather than
// silently dropped.
func Describe(role string) string {
	role = strings.TrimSpace(role)
	if role == "" {
		role = "unknown"
	}
	return role + "/" + Current
}
