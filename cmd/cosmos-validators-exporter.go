package main

import (
	"main/pkg"
	configPkg "main/pkg/config"
	"main/pkg/logger"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "unknown"
)

type OsFS struct {
}

func (fs *OsFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func ExecuteMain(configPath string) {
	filesystem := &OsFS{}
	app := pkg.NewApp(configPath, filesystem, version)
	app.Start()
}

func ExecuteValidateConfig(configPath string) {
	filesystem := &OsFS{}
	config, err := configPkg.GetConfig(configPath, filesystem)
	if err != nil {
		logger.GetDefaultLogger().Panic().Err(err).Msg("Could not load config!")
	}

	if err := config.Validate(); err != nil {
		logger.GetDefaultLogger().Panic().Err(err).Msg("Config is invalid!")
	}

	warnings := config.DisplayWarnings()

	for _, warning := range warnings {
		entry := logger.GetDefaultLogger().Warn()
		for label, value := range warning.Labels {
			entry = entry.Str(label, value)
		}

		entry.Msg(warning.Message)
	}

	logger.GetDefaultLogger().Info().Msg("Provided config is valid.")
}

func main() {
	var ConfigPath string

	rootCmd := &cobra.Command{
		Use:     "cosmos-validators-exporter --config [config path]",
		Long:    "Scrapes validators info on multiple chains.",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			ExecuteMain(ConfigPath)
		},
	}

	validateConfigCmd := &cobra.Command{
		Use:     "validate-config --config [config path]",
		Long:    "Validate config.",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			ExecuteValidateConfig(ConfigPath)
		},
	}

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")
	_ = rootCmd.MarkPersistentFlagRequired("config")

	validateConfigCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")
	_ = validateConfigCmd.MarkPersistentFlagRequired("config")

	rootCmd.AddCommand(validateConfigCmd)

	if err := rootCmd.Execute(); err != nil {
		logger.GetDefaultLogger().Panic().Err(err).Msg("Could not start application")
	}
}
