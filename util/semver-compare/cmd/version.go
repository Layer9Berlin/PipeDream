package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
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

// Cmd implements the version command that outputs information on the current installation as a yaml string map
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
		`name: semver-compare
version: %v
date: %v
commit: %-8v
checksum: %-8v
location: %v
`, Version, Date, CommitHash, RepoChecksum, executableLocation)))
}
