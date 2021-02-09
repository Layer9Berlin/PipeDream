// Package cmd implements the `pipedream` command line tool
package cmd

import (
	"github.com/Layer9Berlin/pipedream/src/logging"
	"github.com/Layer9Berlin/pipedream/src/run"
	"github.com/Layer9Berlin/pipedream/src/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

// RootCmd is the main command that is both the root for subcommands like `version`
// and the default command if no subcommand is specified
// most of the time, you will want the default, which runs applicable pipes
var RootCmd = &cobra.Command{
	Use:   "pipedream",
	Short: "PipeDream",
	Long:  `Layer9 PipeDream - Maintainable script automation`,
	Run:   run.Cmd,
}

func init() {
	cobra.OnInitialize(func() {
		err := logging.SetUpLogs(run.Log, run.Verbosity, os.Stdout)
		if err != nil {
			run.Log.Fatal(err)
		}
	})

	// bind the verbose flag
	// default value is the warn level
	RootCmd.PersistentFlags().StringVarP(&run.Verbosity, "verbosity", "v", logrus.InfoLevel.String(), "Log level (debug, info, warn, error, fatal, panic)")
	RootCmd.PersistentFlags().StringVarP(&run.FileFlag, "file", "f", "", "Path to file containing pipe to execute (default is \"\", ambiguity resolved by user prompt)")
	RootCmd.PersistentFlags().StringVarP(&run.PipelineFlag, "pipe", "p", "", "Identifier of pipeline to execute (default is \"\", ambiguity resolved by user prompt)")

	RootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version of the current PipeDream installation",
		Long:  `All software has versions. This is the current version of your PipeDream installation`,
		Run: func(cmd *cobra.Command, args []string) {
			version.Cmd(cmd.OutOrStdout())
		},
	})
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		run.Log.Fatal(err)
	}
}
