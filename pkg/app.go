package pkg

import (
	fetchersPkg "main/pkg/fetchers"
	generatorsPkg "main/pkg/generators"
	coingeckoPkg "main/pkg/price_fetchers/coingecko"
	dexScreenerPkg "main/pkg/price_fetchers/dex_screener"
	statePkg "main/pkg/state"
	"main/pkg/tracing"
	"main/pkg/types"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"

	"main/pkg/config"
	loggerPkg "main/pkg/logger"
	queriersPkg "main/pkg/queriers"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type App struct {
	Tracer   trace.Tracer
	Config   *config.Config
	Logger   *zerolog.Logger
	Queriers []types.Querier

	Fetchers   []fetchersPkg.Fetcher
	Generators []generatorsPkg.Generator
}

func NewApp(configPath string, version string) *App {
	appConfig, err := config.GetConfig(configPath)
	if err != nil {
		loggerPkg.GetDefaultLogger().Fatal().Err(err).Msg("Could not load config")
	}

	if err = appConfig.Validate(); err != nil {
		loggerPkg.GetDefaultLogger().Fatal().Err(err).Msg("Provided config is invalid!")
	}

	logger := loggerPkg.GetLogger(appConfig.LogConfig)
	appConfig.DisplayWarnings(logger)

	tracer, err := tracing.InitTracer(appConfig.TracingConfig, version)
	if err != nil {
		loggerPkg.GetDefaultLogger().Fatal().Err(err).Msg("Error setting up tracing")
	}

	coingecko := coingeckoPkg.NewCoingecko(appConfig, logger, tracer)
	dexScreener := dexScreenerPkg.NewDexScreener(logger)

	queriers := []types.Querier{
		queriersPkg.NewSelfDelegationsQuerier(logger, appConfig, tracer),
		queriersPkg.NewPriceQuerier(logger, appConfig, tracer, coingecko, dexScreener),
		queriersPkg.NewValidatorQuerier(logger, appConfig, tracer),
	}

	fetchers := []fetchersPkg.Fetcher{
		fetchersPkg.NewSlashingParamsFetcher(logger, appConfig, tracer),
		fetchersPkg.NewSoftOptOutThresholdFetcher(logger, appConfig, tracer),
		fetchersPkg.NewCommissionFetcher(logger, appConfig, tracer),
		fetchersPkg.NewDelegationsFetcher(logger, appConfig, tracer),
		fetchersPkg.NewUnbondsFetcher(logger, appConfig, tracer),
		fetchersPkg.NewSigningInfoFetcher(logger, appConfig, tracer),
		fetchersPkg.NewRewardsFetcher(logger, appConfig, tracer),
		fetchersPkg.NewBalanceFetcher(logger, appConfig, tracer),
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
	}

	return &App{
		Logger:     logger,
		Config:     appConfig,
		Queriers:   queriers,
		Tracer:     tracer,
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

	for _, querierExt := range a.Queriers {
		wg.Add(1)

		go func(querier types.Querier) {
			childQuerierCtx, querierSpan := a.Tracer.Start(
				rootSpanCtx,
				"Querier "+querier.Name(),
				trace.WithAttributes(attribute.String("querier", querier.Name())),
			)
			defer querierSpan.End()

			defer wg.Done()
			collectors, querierQueryInfos := querier.GetMetrics(childQuerierCtx)

			mutex.Lock()
			defer mutex.Unlock()

			for _, collector := range collectors {
				registry.MustRegister(collector)
			}

			queryInfos = append(queryInfos, querierQueryInfos...)
		}(querierExt)
	}

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

	queriesQuerier := queriersPkg.NewQueriesQuerier(a.Config, queryInfos)
	queriesMetrics, _ := queriesQuerier.GetMetrics(rootSpanCtx)

	for _, generator := range a.Generators {
		metrics := generator.Generate(state)
		registry.MustRegister(metrics...)
	}

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
