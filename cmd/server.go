package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wwwil/transponder/pkg/server"
)

var (
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Serve endpoints for each protocol",
		Long:  `Transponder's server creates endpoints for each supported protocol (HTTP, HTTPS, gRPC) on a specified port so that the scanner can connect to it.`,
		Run:   server.Serve,
	}
)

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.PersistentFlags().Uint16Var(
		&server.HTTPPort,
		"http-port",
		8080,
		"The port number to serve HTTP on.",
	)
	serverCmd.PersistentFlags().Uint16Var(
		&server.HTTPSPort,
		"https-port",
		8443,
		"The port number to serve HTTPS on.",
	)
	serverCmd.PersistentFlags().Uint16Var(
		&server.GRPCPort,
		"grpc-port",
		8081,
		"The port number to serve gRPC on.",
	)
}
