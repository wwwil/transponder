package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wwwil/transponder/pkg/scanner"
)

var (
	scannerCmd = &cobra.Command{
		Use:   "scanner",
		Short: "Continuously scan servers",
		Long:  `Transponder's scanner continuously cycles through the list of servers given in the config file and attempts to make a connection on each of the specified ports`,
		Run:   scanner.Scan,
	}
)

func init() {
	rootCmd.AddCommand(scannerCmd)
	scannerCmd.PersistentFlags().StringVarP(
		&scanner.ConfigFilePath,
		"config-file",
		"c",
		"./scanner.yaml",
		"The location of the scanner's config file. The default is `scanner.yaml` in the current directory.",
		)
}
