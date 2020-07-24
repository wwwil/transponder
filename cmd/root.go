package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "transponder",
		Short: "Network connectivity tester",
		Long:  `Transponder is a continuously running multi-protocol network connectivity testing utility for Kubernetes and Istio.`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
