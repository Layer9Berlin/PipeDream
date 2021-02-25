// Package version provides the implementation of the version command, providing information about the current `pipedream` command line tool installation
package version

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Version is the current semantic version (all utils and pipedream itself share the same version)
var Version = "0.0.1"

// CommitHash is the long form of the current build's commit hash
var CommitHash = "-"

// RepoChecksum is a hash of all relevant files in the repo at time of build
//
// This will help tell apart binaries built from source by detecting uncommitted changes.
var RepoChecksum = "-"

// Date is the date and time at which the build was created
var Date = time.Now().Format(time.RFC822)

// Via is the installation method ("npm"/"brew"/"compiled from source"/...)
var Via = "compiled from source"

// Cmd implements the version command that outputs information on the current installation as a yaml string map
func Cmd(writer io.Writer) {
	executableLocation, _ := os.Executable()
	_, _ = writer.Write([]byte(fmt.Sprintf(
		`version: %v
via: %v
date: %v
commit: %-8v
checksum: %-8v
location: %v
`, Version, Via, Date, CommitHash, RepoChecksum, executableLocation)))
}
