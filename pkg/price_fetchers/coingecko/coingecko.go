package coingecko

import (
	"context"
	"fmt"
	"main/pkg/config"
	"main/pkg/http"
	"main/pkg/types"
	"strings"

	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

type Response map[string]map[string]float64

type Coingecko struct {
	Client *http.Client
	Config *config.Config
	Logger zerolog.Logger
	Tracer trace.Tracer
}

func NewCoingecko(
	appConfig *config.Config,
	logger *zerolog.Logger,
	tracer trace.Tracer,
) *Coingecko {
	return &Coingecko{
		Config: appConfig,
		Client: http.NewClient(logger, "coingecko", tracer),
		Logger: logger.With().Str("component", "coingecko").Logger(),
		Tracer: tracer,
	}
}

func (c *Coingecko) FetchPrices(
	currencies []string,
	ctx context.Context,
) (map[string]float64, *types.QueryInfo) {
	childCtx, querierSpan := c.Tracer.Start(
		ctx,
		"Fetching Coingecko prices",
	)
	defer querierSpan.End()

	ids := strings.Join(currencies, ",")
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", ids)

	var response Response
	queryInfo, err := c.Client.Get(url, &response, childCtx)

	if err != nil {
		c.Logger.Error().Err(err).Msg("Could not get rate")
		querierSpan.RecordError(err)
		return nil, &queryInfo
	}

	prices := map[string]float64{}

	for currencyKey, currencyValue := range response {
		for _, baseCurrencyValue := range currencyValue {
			prices[currencyKey] = baseCurrencyValue
		}
	}

	return prices, &queryInfo
}
