package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
)

var version = "0.0.1"
var RepoChecksum = "-"

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version of the current PipeDream installation",
		Long:  `All software has versions. This is the current version of your PipeDream installation`,
		Run: func(cmd *cobra.Command, args []string) {
			versionCmd(cmd.OutOrStdout())
		},
	})
}

func versionCmd(writer io.Writer) {
	_, _ = writer.Write([]byte(fmt.Sprintf("%v (repo checksum: %v)", version, RepoChecksum)))
}
