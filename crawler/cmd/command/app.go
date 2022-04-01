package command

import (
	"context"
	"fmt"
	"github.com/StevenRojas/natscrawler/crawler/config"
	"github.com/StevenRojas/natscrawler/crawler/pkg/crawler"
	"github.com/StevenRojas/natscrawler/crawler/pkg/repository"
	"github.com/spf13/cobra"
)

// NewRootCommand creates the root command
func NewRootCommand(ctx context.Context) *cobra.Command {
	rootCommand := &cobra.Command{
		Use:   "process",
		Short: "Collect URLs and process them",

		PersistentPreRunE: config.Setup,
		RunE: func(cmd *cobra.Command, args []string) error {
			return process(ctx)
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

func process(ctx context.Context) error {
	repo, err := repository.NewRepository(config.App.Database)
	if err != nil {
		return err
	}
	defer func(repo repository.Repository) {
		err := repo.Close()
		if err != nil {
			fmt.Printf("unable to close DB connection: %s", err.Error())
		}
	}(repo)

	c, err := crawler.NewCrawler(repo, config.App)
	if err != nil {
		return err
	}
	c.Process(ctx)
	return nil
}