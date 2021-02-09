package cmd

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"strings"
)

var yamlFlag = false
var ignoredPackages = ""

var log = logrus.New()

var osStdin = os.Stdin
var osStdout = os.Stdout

// DependencyMismatch is a data structure containing information about a package's required and currently installed versions
type DependencyMismatch struct {
	Pm      string
	Package string
	Current string
	Latest  string
	Wanted  string
}

var rootCmd = &cobra.Command{
	Use:   "semver-compare",
	Short: "SemVer Compare",
	Long:  `Compare semantic versions`,
	Run:   runFunction,
}

var runFunction = func(cmd *cobra.Command, args []string) {
	if yamlFlag {
		// we explicitly allow zero-length StdIn input
		stdInText, _ := ioutil.ReadAll(osStdin)
		yamlInput := string(stdInText)

		skipPackage := map[string]bool{}
		if ignoredPackages != "" {
			packageList := strings.Split(ignoredPackages, ",")
			for _, packageName := range packageList {
				skipPackage[packageName] = true
			}
		}
		mismatches := make([]DependencyMismatch, 0, 256)
		err := yaml.Unmarshal([]byte(yamlInput), &mismatches)
		if err != nil {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), err)
			return
		}
		majorConflicts := make([]string, 0, len(mismatches))
		minorConflicts := make([]string, 0, len(mismatches))
		patchConflicts := make([]string, 0, len(mismatches))
		for _, mismatch := range mismatches {
			if skipPackage[mismatch.Package] {
				continue
			}

			currentComponents := strings.Split(mismatch.Current, ".")
			latestComponents := strings.Split(mismatch.Latest, ".")
			if len(currentComponents) < 2 {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "current version `%v` of package `%v` does not seem to be in semantic versioning format\n", mismatch.Current, mismatch.Package)
				return
			}
			if len(currentComponents) == 2 {
				currentComponents = append(currentComponents, "0")
			}
			if len(latestComponents) < 2 {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "latest version `%v` of package `%v` does not seem to be in semantic versioning format\n", mismatch.Latest, mismatch.Package)
				return
			}
			if len(latestComponents) == 2 {
				latestComponents = append(latestComponents, "0")
			}
			if currentComponents[0] != latestComponents[0] {
				majorConflicts = append(majorConflicts, fmt.Sprintf("%v (%v -> %v) %v", mismatch.Package, mismatch.Current, mismatch.Latest, mismatch.Pm))
				continue
			}
			if len(majorConflicts) > 0 {
				continue
			}
			if currentComponents[1] != latestComponents[1] {
				minorConflicts = append(minorConflicts, fmt.Sprintf("%v (%v -> %v) %v", mismatch.Package, mismatch.Current, mismatch.Latest, mismatch.Pm))
				continue
			}
			if len(minorConflicts) > 0 {
				continue
			}
			if currentComponents[2] != latestComponents[2] {
				patchConflicts = append(patchConflicts, fmt.Sprintf("%v (%v -> %v) %v", mismatch.Package, mismatch.Current, mismatch.Latest, mismatch.Pm))
				continue
			}
		}
		if len(majorConflicts) > 0 {
			if len(majorConflicts) > 1 {
				_, _ = fmt.Fprintln(osStdout, "major updates available!")
			} else {
				_, _ = fmt.Fprintln(osStdout, "major updates available!")
			}
			_, _ = fmt.Fprintln(osStdout, strings.Join(majorConflicts, "\n"))
			return
		}
		if len(minorConflicts) > 0 {
			if len(minorConflicts) > 1 {
				_, _ = fmt.Fprintln(osStdout, "minor updates available.")
			} else {
				_, _ = fmt.Fprintln(osStdout, "minor updates available.")
			}
			_, _ = fmt.Fprintln(osStdout, strings.Join(minorConflicts, "\n"))
			return
		}
		if len(patchConflicts) > 0 {
			if len(patchConflicts) > 1 {
				_, _ = fmt.Fprintln(osStdout, "patches available")
			} else {
				_, _ = fmt.Fprintln(osStdout, "patch available")
			}
			_, _ = fmt.Fprintln(osStdout, strings.Join(patchConflicts, "\n"))
			return
		}
		_, _ = fmt.Fprintln(osStdout, "all dependencies up-to-date")
		return
	}
	if len(args) < 2 {
		panic("please provide at least two version arguments")
	}
	components := make([][]string, 0, len(args))
	for _, version := range args {
		versionComponents := strings.Split(version, ".")
		for len(versionComponents) < 3 {
			versionComponents = append(versionComponents, "0")
		}
		components = append(components, versionComponents)
	}
	majorVersion := components[0][0]
	for index, versionComponents := range components {
		if versionComponents[0] != majorVersion {
			_, _ = fmt.Fprintln(osStdout, "major version conflict!")
			_, _ = fmt.Fprintf(osStdout, "%v -> %v\n", args[index], args[0])
			return
		}
	}
	minorVersion := components[0][1]
	for index, versionComponents := range components {
		if versionComponents[1] != minorVersion {
			_, _ = fmt.Fprintln(osStdout, "minor version conflict.")
			_, _ = fmt.Fprintf(osStdout, "%v -> %v\n", args[index], args[0])
			return
		}
	}
	patchVersion := components[0][2]
	for index, versionComponents := range components {
		if versionComponents[2] != patchVersion {
			_, _ = fmt.Fprintln(osStdout, "patch version conflict")
			_, _ = fmt.Fprintf(osStdout, "%v -> %v\n", args[index], args[0])
			return
		}
	}
	_, _ = fmt.Fprintln(osStdout, "versions match")
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&yamlFlag, "yaml", "y", false, "array of dependency mismatches in yaml format")
	rootCmd.PersistentFlags().StringVarP(&ignoredPackages, "ignore-packages", "i", "", "comma-separated list of packages to ignore")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
