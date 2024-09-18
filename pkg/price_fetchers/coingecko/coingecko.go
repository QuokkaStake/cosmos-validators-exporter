package coingecko

import (
	"context"
	"fmt"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/http"
	"main/pkg/price_fetchers"
	"main/pkg/types"
	"main/pkg/utils"
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
	denoms []price_fetchers.ChainWithDenom,
	ctx context.Context,
) ([]price_fetchers.PriceInfo, *types.QueryInfo) {
	childCtx, querierSpan := c.Tracer.Start(
		ctx,
		"Fetching Coingecko prices",
	)
	defer querierSpan.End()

	currencies := utils.Map(denoms, func(c price_fetchers.ChainWithDenom) string {
		return c.DenomInfo.CoingeckoCurrency
	})

	ids := strings.Join(currencies, ",")
	url := fmt.Sprintf(
		"https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=%s",
		ids,
		constants.CoingeckoBaseCurrency,
	)

	var response Response
	queryInfo, _, err := c.Client.Get(url, &response, types.HTTPPredicateAlwaysPass(), childCtx)

	if err != nil {
		c.Logger.Error().Err(err).Msg("Could not get rate")
		querierSpan.RecordError(err)
		return nil, &queryInfo
	}

	pricesInfo := []price_fetchers.PriceInfo{}

	for _, denom := range denoms {
		currency, ok := response[denom.DenomInfo.CoingeckoCurrency]
		if !ok {
			continue
		}

		value, ok := currency[constants.CoingeckoBaseCurrency]
		if !ok {
			continue
		}

		pricesInfo = append(pricesInfo, price_fetchers.PriceInfo{
			Chain:        denom.Chain,
			Denom:        denom.DenomInfo.DisplayDenom,
			BaseCurrency: constants.CoingeckoBaseCurrency,
			Price:        value,
		})
	}

	return pricesInfo, &queryInfo
}
