// Package version holds the build version. The release build sets the vars
// with -ldflags "-X .../internal/version.Version=...".
package version

import "fmt"

// Set by ldflags. The defaults are for a local build.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// String returns the version line of the binary called name.
func String(name string) string {
	return fmt.Sprintf("%s %s (commit %s, built %s)", name, Version, Commit, Date)
}
