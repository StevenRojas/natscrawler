package command

import (
	"context"
	"github.com/StevenRojas/natscrawler/listener/config"
	"github.com/StevenRojas/natscrawler/listener/pkg/listener"
	"github.com/spf13/cobra"
)

// NewRootCommand creates the root command
func NewRootCommand(ctx context.Context) *cobra.Command {
	rootCommand := &cobra.Command{
		Use:   "listen",
		Short: "Listen for URLs to be crawled",

		PersistentPreRunE: config.Setup,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listen()
		},
	}

	rootCommand.Flags().StringVarP(
		&config.Filename,
		"config",
		"c",
		"config/local.yaml",
		"Relative path to the config file",
	)

	return rootCommand
}

func listen() error {
	return listener.NewListener(config.App)
}