package logger

import (
	"os"

	"main/pkg/config"

	"github.com/rs/zerolog"
)

func GetDefaultLogger() *zerolog.Logger {
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	return &log
}

func GetNopLogger() *zerolog.Logger {
	log := zerolog.Nop()
	return &log
}

func GetLogger(appConfig config.LogConfig) *zerolog.Logger {
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	if appConfig.JSONOutput {
		log = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	logLevel, err := zerolog.ParseLevel(appConfig.LogLevel)
	if err != nil {
		log.Panic().Err(err).Msg("Could not parse log level")
	}

	zerolog.SetGlobalLevel(logLevel)
	return &log
}
