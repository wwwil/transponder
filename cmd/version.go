package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wwwil/transponder/pkg/version"
)

var verbose bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the version",
	Long: `Display the Transponder version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.ToString(verbose))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.PersistentFlags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"Whether to display additional information about this Transponder build.",
	)
}

