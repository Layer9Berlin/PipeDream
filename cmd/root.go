package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"pipedream/src/logging"
	"pipedream/src/run"
	"pipedream/src/version"
)

var RootCmd = &cobra.Command{
	Use:   "p",
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
	RootCmd.PersistentFlags().StringVarP(&run.Verbosity, "verbosity", "v", logrus.InfoLevel.String(), "Log level (debug, info, warn, timer, fatal, panic")

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
