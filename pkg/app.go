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
	queriersPkg "main/pkg/queriers"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

type App struct {
	Config   *config.Config
	Logger   *zerolog.Logger
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

	coingecko := coingeckoPkg.NewCoingecko(appConfig, log)
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
		queriersPkg.NewValidatorQuerier(log, appConfig),
		queriersPkg.NewDenomCoefficientsQuerier(log, appConfig),
	}

	return &App{
		Logger:   log,
		Config:   appConfig,
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

	registry := prometheus.NewRegistry()

	var wg sync.WaitGroup
	var mutex sync.Mutex

	var queryInfos []*types.QueryInfo

	for _, querierExt := range a.Queriers {
		wg.Add(1)

		go func(querier types.Querier) {
			defer wg.Done()
			collectors, querierQueryInfos := querier.GetMetrics()

			mutex.Lock()
			defer mutex.Unlock()

			for _, collector := range collectors {
				registry.MustRegister(collector)
			}

			queryInfos = append(queryInfos, querierQueryInfos...)
		}(querierExt)
	}

	wg.Wait()

	queriesQuerier := queriersPkg.NewQueriesQuerier(a.Config, queryInfos)
	queriesMetrics, _ := queriesQuerier.GetMetrics()

	for _, metric := range queriesMetrics {
		registry.MustRegister(metric)
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)

	sublogger.Info().
		Str("method", http.MethodGet).
		Str("endpoint", "/metrics").
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}
