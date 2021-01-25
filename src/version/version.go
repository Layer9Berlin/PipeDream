package version

import (
	"fmt"
	"io"
)

var Version = "0.0.1"
var Commit = ""
var RepoChecksum = "-"
var Date = ""

func Cmd(writer io.Writer) {
	_, _ = writer.Write([]byte(fmt.Sprintf("%v / commit: %v / repo checksum: %v / date: %v", Version, Commit, RepoChecksum, Date)))
}
