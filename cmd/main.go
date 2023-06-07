package main

import (
	"os"
	"path/filepath"

	reposerver "github.com/infor-design/selfservice/cmd/reposerver/commands"
	cli "github.com/infor-design/selfservice/cmd/selfservice/commands"
	server "github.com/infor-design/selfservice/cmd/server/commands"
	wsserver "github.com/infor-design/selfservice/cmd/wsserver/commands"
	"github.com/spf13/cobra"
)

const (
	binaryNameEnv = "BINARY_NAME"
)

func main() {
	var command *cobra.Command

	binaryName := filepath.Base(os.Args[0])
	if val := os.Getenv(binaryNameEnv); val != "" {
		binaryName = val
	}

	switch binaryName {
	case "selfservice", "selfservice-linux-amd64", "selfservice-darwin-amd64", "selfservice-windows-amd64.exe":
		command = cli.NewCommand()
	case "selfservice-server":
		command = server.NewCommand()
	case "selfservice-reposerver":
		command = reposerver.NewCommand()
	case "selfservice-wsserver":
		command = wsserver.NewCommand()
	default:
		command = cli.NewCommand()
	}

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
