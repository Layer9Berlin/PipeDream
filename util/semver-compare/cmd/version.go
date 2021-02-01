package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
)

var Version = "0.0.1"
var CommitHash = "-"
var RepoChecksum = "-"
var Date = "-"
var Via = "compiled from source"

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version of the semver-compare binary",
		Long:  `All software has versions. This is the current version of your semver-compare installation`,
		Run: func(cmd *cobra.Command, args []string) {
			versionCmd(cmd.OutOrStdout())
		},
	})
}

func versionCmd(writer io.Writer) {
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
