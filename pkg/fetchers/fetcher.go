package fetchers

import (
	"context"
	"main/pkg/constants"
	"main/pkg/types"
)

type Fetcher interface {
	Fetch(ctx context.Context) (interface{}, []*types.QueryInfo)
	Name() constants.FetcherName
}
