package pkg

import (
	coingeckoPkg "main/pkg/price_fetchers/coingecko"
	dexScreenerPkg "main/pkg/price_fetchers/dex_screener"
	"main/pkg/types"
	"net/http"
	"sync"
	"time"

	"main/pkg/config"
	"main/pkg/logger"
	managerPkg "main/pkg/manager"
	queriersPkg "main/pkg/queriers"
	"main/pkg/utils"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

type App struct {
	Config   *config.Config
	Logger   *zerolog.Logger
	Manager  *managerPkg.Manager
	Queriers []types.Querier
}

func NewApp(configPath string) *App {
	appConfig, err := config.GetConfig(configPath)
	if err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not load config")
	}

	if err = appConfig.Validate(); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Provided config is invalid!")
	}

	log := logger.GetLogger(appConfig.LogConfig)
	manager := managerPkg.NewManager(appConfig, log)

	coingecko := coingeckoPkg.NewCoingecko(log)
	dexScreener := dexScreenerPkg.NewDexScreener(log)

	queriers := []types.Querier{
		queriersPkg.NewCommissionQuerier(log, appConfig),
		queriersPkg.NewDelegationsQuerier(log, appConfig),
		queriersPkg.NewUnbondsQuerier(log, appConfig),
		queriersPkg.NewSelfDelegationsQuerier(log, appConfig),
		queriersPkg.NewPriceQuerier(log, appConfig, coingecko, dexScreener),
		queriersPkg.NewRewardsQuerier(log, appConfig),
		queriersPkg.NewWalletQuerier(log, appConfig),
		queriersPkg.NewSlashingParamsQuerier(log, appConfig),
	}

	return &App{
		Logger:   log,
		Config:   appConfig,
		Manager:  manager,
		Queriers: queriers,
	}
}

func (a *App) Start() {
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		a.Handler(w, r)
	})

	a.Logger.Info().Str("addr", a.Config.ListenAddress).Msg("Listening")
	err := http.ListenAndServe(a.Config.ListenAddress, nil)
	if err != nil {
		a.Logger.Fatal().Err(err).Msg("Could not start application")
	}
}

func (a *App) Handler(w http.ResponseWriter, r *http.Request) {
	requestStart := time.Now()

	sublogger := a.Logger.With().
		Str("request-id", uuid.New().String()).
		Logger()

	queriesCountGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_queries_total",
			Help: "Total queries done for this chain",
		},
		[]string{"chain", "url"},
	)

	queriesSuccessfulGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_queries_success",
			Help: "Successful queries count for this chain",
		},
		[]string{"chain", "url"},
	)

	queriesFailedGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_queries_error",
			Help: "Failed queries count for this chain",
		},
		[]string{"chain", "url"},
	)

	timingsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_timings",
			Help: "External LCD query timing",
		},
		[]string{"chain", "url"},
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

	missedBlocksGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_missed_blocks",
			Help: "Validator's missed blocks",
		},
		[]string{"chain", "address", "moniker"},
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
	registry.MustRegister(validatorRankGauge)
	registry.MustRegister(validatorsCountGauge)
	registry.MustRegister(missedBlocksGauge)
	registry.MustRegister(activeSetSizeGauge)
	registry.MustRegister(activeSetTokensGauge)
	registry.MustRegister(totalBondedTokensGauge)

	validators := a.Manager.GetAllValidators()
	for _, validator := range validators {
		for _, query := range validator.Queries {
			queriesCountGauge.With(prometheus.Labels{
				"chain": query.Chain,
				"url":   query.URL,
			}).Inc()

			timingsGauge.With(prometheus.Labels{
				"chain": query.Chain,
				"url":   query.URL,
			}).Set(query.Duration.Seconds())

			if query.Success {
				queriesSuccessfulGauge.With(prometheus.Labels{
					"chain": query.Chain,
					"url":   query.URL,
				}).Inc()
			} else {
				queriesFailedGauge.With(prometheus.Labels{
					"chain": query.Chain,
					"url":   query.URL,
				}).Inc()
			}
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
			}).Set(utils.BoolToFloat64(validator.Info.Jailed))

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

		if validator.Info.MissedBlocksCount >= 0 {
			missedBlocksGauge.With(prometheus.Labels{
				"chain":   validator.Chain,
				"address": validator.Address,
				"moniker": validator.Info.Moniker,
			}).Set(float64(validator.Info.MissedBlocksCount))
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

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, querierExt := range a.Queriers {
		wg.Add(1)

		go func(querier types.Querier) {
			defer wg.Done()
			collectors, queryInfos := querier.GetMetrics()

			mutex.Lock()
			defer mutex.Unlock()

			for _, collector := range collectors {
				registry.MustRegister(collector)
			}

			for _, query := range queryInfos {
				queriesCountGauge.With(prometheus.Labels{
					"chain": query.Chain,
					"url":   query.URL,
				}).Inc()

				timingsGauge.With(prometheus.Labels{
					"chain": query.Chain,
					"url":   query.URL,
				}).Set(query.Duration.Seconds())

				if query.Success {
					queriesSuccessfulGauge.With(prometheus.Labels{
						"chain": query.Chain,
						"url":   query.URL,
					}).Inc()
				} else {
					queriesFailedGauge.With(prometheus.Labels{
						"chain": query.Chain,
						"url":   query.URL,
					}).Inc()
				}
			}

		}(querierExt)
	}

	wg.Wait()

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)

	sublogger.Info().
		Str("method", http.MethodGet).
		Str("endpoint", "/metrics").
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}
