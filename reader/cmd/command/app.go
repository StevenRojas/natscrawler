package command

import (
	"context"
	"github.com/StevenRojas/natscrawler/reader/config"
	"github.com/StevenRojas/natscrawler/reader/pkg/parser"
	"github.com/StevenRojas/natscrawler/reader/pkg/sender"
	"github.com/spf13/cobra"
	"sync"
)

// NewRootCommand creates the root command
func NewRootCommand(ctx context.Context) *cobra.Command {
	var csvFile string
	rootCommand := &cobra.Command{
		Use:   "process",
		Short: "Parse and process a CSV file",

		PersistentPreRunE: config.Setup,
		RunE: func(cmd *cobra.Command, args []string) error {
			return process(ctx, csvFile)
		},
	}

	rootCommand.Flags().StringVarP(
		&config.Filename,
		"config",
		"c",
		"config/local.yaml",
		"Relative path to the config file",
	)

	rootCommand.Flags().StringVarP(&csvFile, "file", "f", "", "Path to CSV file")
	rootCommand.MarkFlagRequired("file")

	return rootCommand
}

// process csv file
func process(ctx context.Context, csvFile string) error {
	s, err := sender.NewGrpcService(ctx, config.App.Grpc)
	if err != nil {
		return err
	}
	p, err := parser.NewCSVParser(csvFile, config.App.General.SkipRows, config.App.General.URLDomain)
	if err != nil {
		return err
	}

	urlCh := make(chan string, config.App.General.BufferSize)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go s.Start(ctx, &wg, urlCh)
	go p.Parse(ctx, &wg, urlCh)
	wg.Wait()

	return nil
}