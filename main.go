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

	queriesCountGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_queries_total",
			Help: "Total queries done for this chain",
		},
		[]string{"chain", "address"},
	)

	queriesSuccessfulGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_queries_success",
			Help: "Successful queries count for this chain",
		},
		[]string{"chain", "address"},
	)

	queriesFailedGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_queries_error",
			Help: "Failed queries count for this chain",
		},
		[]string{"chain", "address"},
	)

	timingsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_timings",
			Help: "External LCD query timing",
		},
		[]string{"chain", "address", "url"},
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

	isJailedGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_jailed",
			Help: "Whether a validator is jailed (1 if yes, 0 if no)",
		},
		[]string{"chain", "address", "moniker"},
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

	delegationsCountGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_delegations_count",
			Help: "Validator delegations count",
		},
		[]string{"chain", "address", "moniker"},
	)

	unbondsCountGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_unbonds_count",
			Help: "Validator unbonds count",
		},
		[]string{"chain", "address", "moniker"},
	)

	selfDelegatedTokensGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_self_delegated",
			Help: "Validator's self delegated amount (in tokens)",
		},
		[]string{"chain", "address", "moniker", "denom"},
	)

	validatorRankGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_rank",
			Help: "Rank of a validator compared to other validators on chain.",
		},
		[]string{"chain", "address", "moniker"},
	)

	validatorsCountGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_validators_count",
			Help: "Total active validators count on chain.",
		},
		[]string{"chain", "address", "moniker"},
	)

	commissionUnclaimedTokens := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_unclaimed_commission",
			Help: "Validator's unclaimed commission (in tokens)",
		},
		[]string{"chain", "address", "moniker", "denom"},
	)

	selfDelegationRewardsTokens := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_self_delegation_rewards",
			Help: "Validator's self-delegation rewards (in tokens)",
		},
		[]string{"chain", "address", "moniker", "denom"},
	)

	walletBalanceTokens := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_wallet_balance",
			Help: "Validator's wallet balance (in tokens)",
		},
		[]string{"chain", "address", "moniker", "denom"},
	)

	missedBlocksGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_missed_blocks",
			Help: "Validator's missed blocks",
		},
		[]string{"chain", "address", "moniker"},
	)

	blocksWindowGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_missed_blocks_window",
			Help: "Missed blocks window in network",
		},
		[]string{"chain"},
	)

	denomCoefficientGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_denom_coefficient",
			Help: "Denom coefficient info",
		},
		[]string{"chain", "denom", "display_denom"},
	)

	activeSetSizeGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_active_set_size",
			Help: "Active set size",
		},
		[]string{"chain"},
	)

	activeSetTokensGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_active_set_tokens",
			Help: "Tokens needed to get into active set (last validators' stake, or 0 if not enough validators)",
		},
		[]string{"chain"},
	)

	tokenPriceGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_price",
			Help: "Price of 1 token in display denom in USD",
		},
		[]string{"chain"},
	)

	totalBondedTokensGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_tokens_bonded_total",
			Help: "Total tokens bonded in chain",
		},
		[]string{"chain"},
	)

	registry := prometheus.NewRegistry()
	registry.MustRegister(queriesCountGauge)
	registry.MustRegister(queriesSuccessfulGauge)
	registry.MustRegister(queriesFailedGauge)
	registry.MustRegister(timingsGauge)
	registry.MustRegister(validatorInfoGauge)
	registry.MustRegister(isJailedGauge)
	registry.MustRegister(commissionGauge)
	registry.MustRegister(commissionMaxGauge)
	registry.MustRegister(commissionMaxChangeGauge)
	registry.MustRegister(delegationsGauge)
	registry.MustRegister(unbondsCountGauge)
	registry.MustRegister(delegationsCountGauge)
	registry.MustRegister(selfDelegatedTokensGauge)
	registry.MustRegister(validatorRankGauge)
	registry.MustRegister(validatorsCountGauge)
	registry.MustRegister(commissionUnclaimedTokens)
	registry.MustRegister(selfDelegationRewardsTokens)
	registry.MustRegister(walletBalanceTokens)
	registry.MustRegister(missedBlocksGauge)
	registry.MustRegister(blocksWindowGauge)
	registry.MustRegister(denomCoefficientGauge)
	registry.MustRegister(activeSetSizeGauge)
	registry.MustRegister(activeSetTokensGauge)
	registry.MustRegister(tokenPriceGauge)
	registry.MustRegister(totalBondedTokensGauge)

	validators := manager.GetAllValidators()
	for _, validator := range validators {
		queriesCountGauge.With(prometheus.Labels{
			"chain":   validator.Chain,
			"address": validator.Address,
		}).Set(float64(len(validator.Queries)))

		queriesSuccessfulGauge.With(prometheus.Labels{
			"chain":   validator.Chain,
			"address": validator.Address,
		}).Set(validator.GetSuccessfulQueriesCount())

		queriesFailedGauge.With(prometheus.Labels{
			"chain":   validator.Chain,
			"address": validator.Address,
		}).Set(validator.GetFailedQueriesCount())

		for _, query := range validator.Queries {
			timingsGauge.With(prometheus.Labels{
				"chain":   validator.Chain,
				"address": validator.Address,
				"url":     query.URL,
			}).Set(query.Duration.Seconds())
		}

		if validator.Info.Moniker != "" { // validator request may fail, here it's assumed it didn't
			validatorInfoGauge.With(prometheus.Labels{
				"chain":            validator.Chain,
				"address":          validator.Address,
				"moniker":          validator.Info.Moniker,
				"details":          validator.Info.Details,
				"identity":         validator.Info.Identity,
				"security_contact": validator.Info.SecurityContact,
				"website":          validator.Info.Website,
			}).Set(1)

			isJailedGauge.With(prometheus.Labels{
				"chain":   validator.Chain,
				"address": validator.Address,
				"moniker": validator.Info.Moniker,
			}).Set(BoolToFloat64(validator.Info.Jailed))

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
		}

		if validator.Info.DelegatorsCount != 0 {
			delegationsCountGauge.With(prometheus.Labels{
				"chain":   validator.Chain,
				"address": validator.Address,
				"moniker": validator.Info.Moniker,
			}).Set(float64(validator.Info.DelegatorsCount))
		}

		if validator.Info.UnbondsCount != -1 {
			unbondsCountGauge.With(prometheus.Labels{
				"chain":   validator.Chain,
				"address": validator.Address,
				"moniker": validator.Info.Moniker,
			}).Set(float64(validator.Info.UnbondsCount))
		}

		if validator.Info.SelfDelegation.Amount != 0 {
			selfDelegatedTokensGauge.With(prometheus.Labels{
				"chain":   validator.Chain,
				"address": validator.Address,
				"moniker": validator.Info.Moniker,
				"denom":   validator.Info.SelfDelegation.Denom,
			}).Set(validator.Info.SelfDelegation.Amount)
		}

		if validator.Info.Rank != 0 {
			validatorRankGauge.With(prometheus.Labels{
				"chain":   validator.Chain,
				"address": validator.Address,
				"moniker": validator.Info.Moniker,
			}).Set(float64(validator.Info.Rank))
		}

		if validator.Info.TotalValidators != -1 {
			validatorsCountGauge.With(prometheus.Labels{
				"chain":   validator.Chain,
				"address": validator.Address,
				"moniker": validator.Info.Moniker,
			}).Set(float64(validator.Info.TotalValidators))
		}

		for _, balance := range validator.Info.Commission {
			commissionUnclaimedTokens.With(prometheus.Labels{
				"chain":   validator.Chain,
				"address": validator.Address,
				"moniker": validator.Info.Moniker,
				"denom":   balance.Denom,
			}).Set(balance.Amount)
		}

		for _, balance := range validator.Info.SelfDelegationRewards {
			selfDelegationRewardsTokens.With(prometheus.Labels{
				"chain":   validator.Chain,
				"address": validator.Address,
				"moniker": validator.Info.Moniker,
				"denom":   balance.Denom,
			}).Set(balance.Amount)
		}

		for _, balance := range validator.Info.WalletBalance {
			walletBalanceTokens.With(prometheus.Labels{
				"chain":   validator.Chain,
				"address": validator.Address,
				"moniker": validator.Info.Moniker,
				"denom":   balance.Denom,
			}).Set(balance.Amount)
		}

		if validator.Info.MissedBlocksCount >= 0 {
			missedBlocksGauge.With(prometheus.Labels{
				"chain":   validator.Chain,
				"address": validator.Address,
				"moniker": validator.Info.Moniker,
			}).Set(float64(validator.Info.MissedBlocksCount))
		}

		if validator.Info.SignedBlocksWindow > 0 {
			blocksWindowGauge.With(prometheus.Labels{
				"chain": validator.Chain,
			}).Set(float64(validator.Info.SignedBlocksWindow))
		}

		if validator.Info.ActiveValidatorsCount >= 0 {
			activeSetSizeGauge.With(prometheus.Labels{
				"chain": validator.Chain,
			}).Set(float64(validator.Info.ActiveValidatorsCount))
		}

		activeSetTokensGauge.With(prometheus.Labels{
			"chain": validator.Chain,
		}).Set(validator.Info.LastValidatorStake)

		totalBondedTokensGauge.With(prometheus.Labels{
			"chain": validator.Chain,
		}).Set(validator.Info.TotalStake)
	}

	for _, chain := range manager.Config.Chains {
		denomCoefficientGauge.With(prometheus.Labels{
			"chain":         chain.Name,
			"display_denom": chain.Denom,
			"denom":         chain.BaseDenom,
		}).Set(float64(chain.DenomCoefficient))
	}

	currencies := manager.GetCurrencies()
	for chain, price := range currencies {
		tokenPriceGauge.With(prometheus.Labels{
			"chain": chain,
		}).Set(price)
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)

	sublogger.Info().
		Str("method", http.MethodGet).
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
