package main

import (
	"main/pkg"
	"main/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	version = "unknown"
)

func Execute(configPath string) {
	app := pkg.NewApp(configPath)
	app.Start()
}

func main() {
	var ConfigPath string

	rootCmd := &cobra.Command{
		Use:     "cosmos-validators-exporter",
		Long:    "Scrapes validators info on multiple chains.",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			Execute(ConfigPath)
		},
	}

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")
	if err := rootCmd.MarkPersistentFlagRequired("config"); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not set flag as required")
	}

	if err := rootCmd.Execute(); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not start application")
	}
}
