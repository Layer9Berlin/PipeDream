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
commit: %v
checksum: %v
date: %v
via: %v
location: %v
`, Version, CommitHash, RepoChecksum, Date, Via, executableLocation)))
}
