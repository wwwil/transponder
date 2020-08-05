package version

import (
	"fmt"
)

// These variables are injected from the Makefile at build time.

// TransponderVersion hosts the version of the app.
var TransponderVersion = "development"

// Commit is the commit hash of the build.
var Commit string

// BuildDate is the date of the build.
var BuildDate string

// GoVersion is the Go version used for the build.
var GoVersion string

// Platform is the target platform for this build.
var Platform string

func ToString(verbose bool) string {
	versionString := fmt.Sprintf("Transponder version: %s %s", TransponderVersion, Platform)
	if verbose {
		versionString = fmt.Sprintf("%s\n  Commit: %s", versionString, Commit)
		versionString = fmt.Sprintf("%s\n  Built: %s", versionString, BuildDate)
		versionString = fmt.Sprintf("%s\n  Go: %s", versionString, GoVersion)
	}
	return versionString
}