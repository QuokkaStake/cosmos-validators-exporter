package pkg

import (
	coingeckoPkg "main/pkg/price_fetchers/coingecko"
	dexScreenerPkg "main/pkg/price_fetchers/dex_screener"
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
		queriersPkg.NewCommissionQuerier(logger, appConfig, tracer),
		queriersPkg.NewDelegationsQuerier(logger, appConfig, tracer),
		queriersPkg.NewUnbondsQuerier(logger, appConfig, tracer),
		queriersPkg.NewSelfDelegationsQuerier(logger, appConfig, tracer),
		queriersPkg.NewPriceQuerier(logger, appConfig, tracer, coingecko, dexScreener),
		queriersPkg.NewRewardsQuerier(logger, appConfig, tracer),
		queriersPkg.NewWalletQuerier(logger, appConfig, tracer),
		queriersPkg.NewSlashingParamsQuerier(logger, appConfig, tracer),
		queriersPkg.NewValidatorQuerier(logger, appConfig, tracer),
		queriersPkg.NewDenomCoefficientsQuerier(logger, appConfig),
		queriersPkg.NewSigningInfoQuerier(logger, appConfig, tracer),
		queriersPkg.NewChainInfoQuerier(logger, appConfig, tracer),
		queriersPkg.NewUptimeQuerier(),
	}

	return &App{
		Logger:   logger,
		Config:   appConfig,
		Queriers: queriers,
		Tracer:   tracer,
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

	wg.Wait()

	queriesQuerier := queriersPkg.NewQueriesQuerier(a.Config, queryInfos)
	queriesMetrics, _ := queriesQuerier.GetMetrics(rootSpanCtx)

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
