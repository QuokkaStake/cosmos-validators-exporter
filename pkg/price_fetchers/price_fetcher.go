package price_fetchers

import (
	"context"
	configPkg "main/pkg/config"
	"main/pkg/types"
)

type ChainWithDenom struct {
	Chain     string
	DenomInfo *configPkg.DenomInfo
}

type PriceInfo struct {
	Chain        string
	Denom        string
	BaseCurrency string
	Price        float64
}

type PriceFetcher interface {
	FetchPrices(denoms []ChainWithDenom, ctx context.Context) ([]PriceInfo, *types.QueryInfo)
}
