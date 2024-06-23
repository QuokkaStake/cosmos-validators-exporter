package pkg

import (
	fetchersPkg "main/pkg/fetchers"
	"main/pkg/fs"
	generatorsPkg "main/pkg/generators"
	coingeckoPkg "main/pkg/price_fetchers/coingecko"
	dexScreenerPkg "main/pkg/price_fetchers/dex_screener"
	statePkg "main/pkg/state"
	"main/pkg/tendermint"
	"main/pkg/tracing"
	"main/pkg/types"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"

	"main/pkg/config"
	loggerPkg "main/pkg/logger"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type App struct {
	Tracer trace.Tracer
	Config *config.Config
	Logger *zerolog.Logger

	RPCs map[string]*tendermint.RPCWithConsumers

	// Fetcher is a class that fetch data and is later stored in state.
	// It doesn't provide any metrics, only data to generate them later.
	Fetchers []fetchersPkg.Fetcher

	// Generator is a class that takes some metrics from the state
	// that were fetcher by one or more Fetchers and generates one or more
	// metrics based on this data.
	// Example: ActiveSetTokenGenerator generates a metric
	// based on ValidatorsFetcher and StakingParamsFetcher.
	Generators []generatorsPkg.Generator
}

func NewApp(configPath string, filesystem fs.FS, version string) *App {
	appConfig, err := config.GetConfig(configPath, filesystem)
	if err != nil {
		loggerPkg.GetDefaultLogger().Fatal().Err(err).Msg("Could not load config")
	}

	if err = appConfig.Validate(); err != nil {
		loggerPkg.GetDefaultLogger().Fatal().Err(err).Msg("Provided config is invalid!")
	}

	logger := loggerPkg.GetLogger(appConfig.LogConfig)
	warnings := appConfig.DisplayWarnings()
	for _, warning := range warnings {
		entry := logger.Warn()
		for label, value := range warning.Labels {
			entry = entry.Str(label, value)
		}

		entry.Msg(warning.Message)
	}

	tracer, err := tracing.InitTracer(appConfig.TracingConfig, version)
	if err != nil {
		loggerPkg.GetDefaultLogger().Fatal().Err(err).Msg("Error setting up tracing")
	}

	coingecko := coingeckoPkg.NewCoingecko(appConfig, logger, tracer)
	dexScreener := dexScreenerPkg.NewDexScreener(logger)

	rpcs := make(map[string]*tendermint.RPCWithConsumers, len(appConfig.Chains))

	for _, chain := range appConfig.Chains {
		rpcs[chain.Name] = tendermint.RPCWithConsumersFromChain(chain, appConfig.Timeout, *logger, tracer)
	}

	fetchers := []fetchersPkg.Fetcher{
		fetchersPkg.NewSlashingParamsFetcher(logger, appConfig, rpcs, tracer),
		fetchersPkg.NewSoftOptOutThresholdFetcher(logger, appConfig, rpcs, tracer),
		fetchersPkg.NewCommissionFetcher(logger, appConfig, rpcs, tracer),
		fetchersPkg.NewDelegationsFetcher(logger, appConfig, rpcs, tracer),
		fetchersPkg.NewUnbondsFetcher(logger, appConfig, rpcs, tracer),
		fetchersPkg.NewSigningInfoFetcher(logger, appConfig, rpcs, tracer),
		fetchersPkg.NewRewardsFetcher(logger, appConfig, rpcs, tracer),
		fetchersPkg.NewBalanceFetcher(logger, appConfig.Chains, rpcs, tracer),
		fetchersPkg.NewSelfDelegationFetcher(logger, appConfig, rpcs, tracer),
		fetchersPkg.NewValidatorsFetcher(logger, appConfig, rpcs, tracer),
		fetchersPkg.NewConsumerValidatorsFetcher(logger, appConfig, rpcs, tracer),
		fetchersPkg.NewStakingParamsFetcher(logger, appConfig, rpcs, tracer),
		fetchersPkg.NewPriceFetcher(logger, appConfig, tracer, coingecko, dexScreener),
		fetchersPkg.NewNodeInfoFetcher(logger, appConfig, rpcs, tracer),
		fetchersPkg.NewConsumerInfoFetcher(logger, appConfig, rpcs, tracer),
	}

	generators := []generatorsPkg.Generator{
		generatorsPkg.NewSlashingParamsGenerator(),
		generatorsPkg.NewSoftOptOutThresholdGenerator(),
		generatorsPkg.NewIsConsumerGenerator(appConfig.Chains),
		generatorsPkg.NewDenomCoefficientGenerator(appConfig.Chains),
		generatorsPkg.NewUptimeGenerator(),
		generatorsPkg.NewCommissionGenerator(),
		generatorsPkg.NewDelegationsGenerator(),
		generatorsPkg.NewUnbondsGenerator(),
		generatorsPkg.NewSigningInfoGenerator(),
		generatorsPkg.NewRewardsGenerator(),
		generatorsPkg.NewBalanceGenerator(),
		generatorsPkg.NewSelfDelegationGenerator(),
		generatorsPkg.NewValidatorsInfoGenerator(),
		generatorsPkg.NewSingleValidatorInfoGenerator(appConfig.Chains, logger),
		generatorsPkg.NewValidatorRankGenerator(appConfig.Chains, logger),
		generatorsPkg.NewActiveSetTokensGenerator(appConfig.Chains),
		generatorsPkg.NewNodeInfoGenerator(),
		generatorsPkg.NewStakingParamsGenerator(),
		generatorsPkg.NewPriceGenerator(),
		generatorsPkg.NewConsumerInfoGenerator(),
	}

	return &App{
		Logger:     logger,
		Config:     appConfig,
		Tracer:     tracer,
		RPCs:       rpcs,
		Fetchers:   fetchers,
		Generators: generators,
	}
}

func (a *App) Start() {
	otelHandler := otelhttp.NewHandler(http.HandlerFunc(a.Handler), "prometheus")
	http.Handle("/metrics", otelHandler)

	a.Logger.Info().Str("addr", a.Config.ListenAddress).Msg("Listening")
	err := http.ListenAndServe(a.Config.ListenAddress, nil)
	if err != nil {
		a.Logger.Fatal().Err(err).Msg("Could not start application")
	}
}

func (a *App) Handler(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()

	span := trace.SpanFromContext(r.Context())
	span.SetAttributes(attribute.String("request-id", requestID))
	rootSpanCtx := r.Context()

	defer span.End()

	requestStart := time.Now()

	sublogger := a.Logger.With().
		Str("request-id", requestID).
		Logger()

	registry := prometheus.NewRegistry()

	var wg sync.WaitGroup
	var mutex sync.Mutex

	var queryInfos []*types.QueryInfo

	state := statePkg.NewState()

	for _, fetchersExt := range a.Fetchers {
		wg.Add(1)

		go func(fetcher fetchersPkg.Fetcher) {
			childQuerierCtx, fetcherSpan := a.Tracer.Start(
				rootSpanCtx,
				"Fetcher "+string(fetcher.Name()),
				trace.WithAttributes(attribute.String("fetcher", string(fetcher.Name()))),
			)
			defer fetcherSpan.End()

			defer wg.Done()
			data, fetcherQueryInfos := fetcher.Fetch(childQuerierCtx)

			mutex.Lock()
			state.Set(fetcher.Name(), data)
			queryInfos = append(queryInfos, fetcherQueryInfos...)
			mutex.Unlock()
		}(fetchersExt)
	}

	wg.Wait()

	queriesMetrics := NewQueriesMetrics(a.Config, queryInfos)
	registry.MustRegister(queriesMetrics.GetMetrics(rootSpanCtx)...)

	for _, generator := range a.Generators {
		metrics := generator.Generate(state)
		registry.MustRegister(metrics...)
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)

	sublogger.Info().
		Str("method", http.MethodGet).
		Str("endpoint", "/metrics").
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}
