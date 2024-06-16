package fetchers

import (
	"context"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/tendermint"
	"main/pkg/types"
	"main/pkg/utils"
	"sync"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type BalanceFetcher struct {
	Logger zerolog.Logger
	Config *config.Config
	RPCs   map[string]*tendermint.RPCWithConsumers
	Tracer trace.Tracer

	wg    sync.WaitGroup
	mutex sync.Mutex

	queryInfos  []*types.QueryInfo
	allBalances map[string]map[string][]types.Amount
}

type BalanceData struct {
	Balances map[string]map[string][]types.Amount
}

func NewBalanceFetcher(
	logger *zerolog.Logger,
	config *config.Config,
	rpcs map[string]*tendermint.RPCWithConsumers,
	tracer trace.Tracer,
) *BalanceFetcher {
	return &BalanceFetcher{
		Logger: logger.With().Str("component", "balance_fetcher").Logger(),
		Config: config,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (q *BalanceFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
	q.queryInfos = []*types.QueryInfo{}
	q.allBalances = map[string]map[string][]types.Amount{}

	for _, chain := range q.Config.Chains {
		q.allBalances[chain.Name] = map[string][]types.Amount{}
		for _, consumerChain := range chain.ConsumerChains {
			q.allBalances[consumerChain.Name] = map[string][]types.Amount{}
		}
	}

	for _, chain := range q.Config.Chains {
		rpc, _ := q.RPCs[chain.Name]

		for _, validator := range chain.Validators {
			q.wg.Add(1 + len(chain.ConsumerChains))

			go q.processChain(
				ctx,
				chain.Name,
				chain.BechWalletPrefix,
				validator.Address,
				rpc.RPC,
			)

			for consumerIndex, consumerChain := range chain.ConsumerChains {
				consumerRPC := rpc.Consumers[consumerIndex]

				go q.processChain(
					ctx,
					consumerChain.Name,
					consumerChain.BechWalletPrefix,
					validator.Address,
					consumerRPC,
				)
			}
		}
	}

	q.wg.Wait()

	return BalanceData{Balances: q.allBalances}, q.queryInfos
}

func (q *BalanceFetcher) Name() constants.FetcherName {
	return constants.FetcherNameBalance
}

func (q *BalanceFetcher) processChain(
	ctx context.Context,
	chainName string,
	chainBechWalletPrefix string,
	validator string,
	rpc *tendermint.RPC,
) {
	defer q.wg.Done()

	if chainBechWalletPrefix == "" {
		return
	}

	wallet, err := utils.ChangeBech32Prefix(validator, chainBechWalletPrefix)
	if err != nil {
		q.Logger.Error().
			Err(err).
			Str("chain", chainName).
			Str("address", validator).
			Msg("Error converting validator address")
		return
	}

	balances, query, err := rpc.GetWalletBalance(wallet, ctx)

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if query != nil {
		q.queryInfos = append(q.queryInfos, query)
	}

	if err != nil {
		q.Logger.Error().
			Err(err).
			Str("chain", chainName).
			Str("address", validator).
			Msg("Error querying for validator wallet balance")
		return
	}

	if balances == nil {
		return
	}

	q.allBalances[chainName][validator] = balances
}
