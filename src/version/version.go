package version

import (
	"fmt"
	"io"
)

var version = "0.0.1"
var RepoChecksum = "-"

func Cmd(writer io.Writer) {
	_, _ = writer.Write([]byte(fmt.Sprintf("%v (repo checksum: %v)", version, RepoChecksum)))
}
