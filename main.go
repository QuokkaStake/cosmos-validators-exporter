package main

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func Execute(configPath string) {
	config, err := GetConfig(configPath)
	if err != nil {
		GetDefaultLogger().Fatal().Err(err).Msg("Could not load config")
	}

	if err = config.Validate(); err != nil {
		GetDefaultLogger().Fatal().Err(err).Msg("Provided config is invalid!")
	}

	log := GetLogger(config.LogConfig)
	manager := NewManager(*config, log)

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		Handler(w, r, manager, log)
	})

	log.Info().Str("addr", config.ListenAddress).Msg("Listening")
	err = http.ListenAndServe(config.ListenAddress, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not start application")
	}
}

func Handler(w http.ResponseWriter, r *http.Request, manager *Manager, log *zerolog.Logger) {
	requestStart := time.Now()

	sublogger := log.With().
		Str("request-id", uuid.New().String()).
		Logger()

	successGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_success",
			Help: "Whether a scrape was successful",
		},
		[]string{"chain", "address"},
	)

	timingsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_timings",
			Help: "External LCD query timing",
		},
		[]string{"chain", "address"},
	)

	validatorInfoGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_info",
			Help: "Validator info",
		},
		[]string{
			"chain",
			"address",
			"moniker",
			"details",
			"identity",
			"security_contact",
			"website",
		},
	)

	commissionGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_commission",
			Help: "Validator current commission",
		},
		[]string{"chain", "address", "moniker"},
	)

	commissionMaxGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_commission_max",
			Help: "Max commission for validator",
		},
		[]string{"chain", "address", "moniker"},
	)

	commissionMaxChangeGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_commission_max_change",
			Help: "Max commission change for validator",
		},
		[]string{"chain", "address", "moniker"},
	)

	delegationsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_total_delegations",
			Help: "Validator delegations (in tokens)",
		},
		[]string{"chain", "address", "moniker"},
	)

	delegationsUsdGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_total_delegations_usd",
			Help: "Validator delegations (in USD)",
		},
		[]string{"chain", "address", "moniker"},
	)

	registry := prometheus.NewRegistry()
	registry.MustRegister(successGauge)
	registry.MustRegister(timingsGauge)
	registry.MustRegister(validatorInfoGauge)
	registry.MustRegister(commissionGauge)
	registry.MustRegister(commissionMaxGauge)
	registry.MustRegister(commissionMaxChangeGauge)
	registry.MustRegister(delegationsGauge)
	registry.MustRegister(delegationsUsdGauge)

	validators := manager.GetAllValidators()
	for _, validator := range validators {
		successGauge.With(prometheus.Labels{
			"chain":   validator.Chain,
			"address": validator.Address,
		}).Set(BoolToFloat64(validator.Success))

		timingsGauge.With(prometheus.Labels{
			"chain":   validator.Chain,
			"address": validator.Address,
		}).Set(validator.Duration.Seconds())

		if !validator.Success {
			continue
		}

		validatorInfoGauge.With(prometheus.Labels{
			"chain":            validator.Chain,
			"address":          validator.Address,
			"moniker":          validator.Info.Moniker,
			"details":          validator.Info.Details,
			"identity":         validator.Info.Identity,
			"security_contact": validator.Info.SecurityContact,
			"website":          validator.Info.Website,
		}).Set(1)

		commissionGauge.With(prometheus.Labels{
			"chain":   validator.Chain,
			"address": validator.Address,
			"moniker": validator.Info.Moniker,
		}).Set(validator.Info.CommissionRate)

		commissionMaxGauge.With(prometheus.Labels{
			"chain":   validator.Chain,
			"address": validator.Address,
			"moniker": validator.Info.Moniker,
		}).Set(validator.Info.CommissionMaxRate)

		commissionMaxChangeGauge.With(prometheus.Labels{
			"chain":   validator.Chain,
			"address": validator.Address,
			"moniker": validator.Info.Moniker,
		}).Set(validator.Info.CommissionMaxChangeRate)

		delegationsGauge.With(prometheus.Labels{
			"chain":   validator.Chain,
			"address": validator.Address,
			"moniker": validator.Info.Moniker,
		}).Set(validator.Info.Tokens)

		if validator.Info.TokensUSD != 0 {
			delegationsUsdGauge.With(prometheus.Labels{
				"chain":   validator.Chain,
				"address": validator.Address,
				"moniker": validator.Info.Moniker,
			}).Set(validator.Info.TokensUSD)
		}
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)

	sublogger.Info().
		Str("method", "GET").
		Str("endpoint", "/metrics").
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}

func main() {
	var ConfigPath string

	rootCmd := &cobra.Command{
		Use:  "cosmos-validators-exporter",
		Long: "Scrapes validators info on multiple chains.",
		Run: func(cmd *cobra.Command, args []string) {
			Execute(ConfigPath)
		},
	}

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")
	if err := rootCmd.MarkPersistentFlagRequired("config"); err != nil {
		GetDefaultLogger().Fatal().Err(err).Msg("Could not set flag as required")
	}

	if err := rootCmd.Execute(); err != nil {
		GetDefaultLogger().Fatal().Err(err).Msg("Could not start application")
	}
}
