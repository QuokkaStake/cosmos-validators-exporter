package fetchers

import (
	"context"
	"main/pkg/constants"
	"main/pkg/types"
)

type StubFetcher1 struct{}

func (f *StubFetcher1) Name() constants.FetcherName {
	return constants.FetcherNameStub1
}

func (f *StubFetcher1) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{}
}

func (f *StubFetcher1) Fetch(
	ctx context.Context,
	data ...interface{},
) (interface{}, []*types.QueryInfo) {
	return nil, []*types.QueryInfo{}
}

type StubFetcher2 struct{}

func (f *StubFetcher2) Name() constants.FetcherName {
	return constants.FetcherNameStub2
}

func (f *StubFetcher2) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{constants.FetcherNameStub1}
}

func (f *StubFetcher2) Fetch(
	ctx context.Context,
	data ...interface{},
) (interface{}, []*types.QueryInfo) {
	return nil, []*types.QueryInfo{}
}
