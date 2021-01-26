package cmd

import (
	"bufio"
	"fmt"
	"github.com/logrusorgru/aurora/v3"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

var redRegex = ""
var amberRegex = ""
var greenRegex = ""
var text = ""
var prefix = ""

var rootCmd = &cobra.Command{
	Use:   "traffic-light",
	Short: "Traffic Light Status Indicator",
	Long:  `Simple status indicator showing some text and a red/amber/green indicator`,
	Run: func(cmd *cobra.Command, args []string) {
		if !cmd.PersistentFlags().Lookup("text").Changed {
			stdInReader := bufio.NewReader(os.Stdin)
			stdInText, err := ioutil.ReadAll(stdInReader)
			if err != nil {
				fmt.Println(aurora.Red(fmt.Sprint("[✘]", fmt.Errorf(" failed to read StdIn, try using the `--text` option instead: %w", err))))
				return
			}
			text = string(stdInText)
		}

		formattedPrefix := ""
		if prefix != "" {
			if strings.HasSuffix(strings.Trim(prefix, " "), ":") {
				formattedPrefix = prefix
			} else {
				formattedPrefix = prefix + ": "
			}
		}
		if redRegex != "" {
			compiledRedRegex := regexp.MustCompile(redRegex)
			if compiledRedRegex.MatchString(text) {
				fmt.Println(aurora.Red(fmt.Sprint("[✘]", " ", formattedPrefix, text)))
				return
			}
		}
		if amberRegex != "" {
			compiledAmberRegex := regexp.MustCompile(amberRegex)
			if compiledAmberRegex.MatchString(text) {
				fmt.Println(aurora.Yellow(fmt.Sprint("[-]", " ", formattedPrefix, text)))
				return
			}
		}
		if greenRegex != "" {
			compiledGreenRegex := regexp.MustCompile(greenRegex)
			if !compiledGreenRegex.MatchString(text) {
				fmt.Println(aurora.Red(fmt.Sprint("[✘]", " ", formattedPrefix, text)))
				return
			}
		}

		fmt.Println(aurora.Green("[✔]"), aurora.Gray(18, fmt.Sprint(formattedPrefix, text)))
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&redRegex, "red", "r", "", "regex to match for red status")
	rootCmd.PersistentFlags().StringVarP(&amberRegex, "amber", "a", "", "regex to match for amber status")
	rootCmd.PersistentFlags().StringVarP(&greenRegex, "green", "g", "", "regex to match for green status")

	rootCmd.PersistentFlags().StringVarP(&text, "text", "t", "", "status text to display")
	rootCmd.PersistentFlags().StringVarP(&prefix, "prefix", "p", "", "status text prefix")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
