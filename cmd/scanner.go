package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wwwil/transponder/pkg/scanner"
)

var (
	scannerCmd = &cobra.Command{
		Use:   "scanner",
		Short: "Transponder scanner component",
		Long:  `Transponder is a continuously running multi-protocol network connectivity testing utility for Kubernetes and Istio.`,
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
