package commands

import (
	"github.com/infor-design/selfservice/server"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:               "selfservice-server",
		Short:             "Run the selfservice API server",
		Long:              "The API server is a REST server which exposes the API consumed by the Web UI, and CLI.  This command runs API server in the foreground.  It can be configured by following options.",
		DisableAutoGenTag: true,
		Run: func(c *cobra.Command, args []string) {
			serverConfig := server.ServerConfig{}
			server := server.NewServer(serverConfig)
			server.Init()
			server.Run()
		},
	}

	return command
}
