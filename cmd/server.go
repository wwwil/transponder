package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wwwil/transponder/pkg/server"
)

var (
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Transponder server component",
		Long:  `Transponder is a continuously running multi-protocol network connectivity testing utility for Kubernetes and Istio.`,
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
