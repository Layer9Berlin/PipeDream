// Implementation of the version command, providing information about the current pipedream command line tool installation
package version

import (
	"fmt"
	"io"
	"os"
)

var Version = "0.0.1"
var CommitHash = "-"
var RepoChecksum = "-"
var Date = "-"
var Via = "compiled from source"

func Cmd(writer io.Writer) {
	executableLocation, _ := os.Executable()
	_, _ = writer.Write([]byte(fmt.Sprintf(
		`version: %v
via: %v
date: %v
commit: %v
checksum: %v
location: %v
`, Version, Via, Date, CommitHash, RepoChecksum, executableLocation)))
}
