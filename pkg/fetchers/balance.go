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
	var queryInfos []*types.QueryInfo

	allBalances := map[string]map[string][]types.Amount{}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	processChain := func(
		chainName string,
		chainBechWalletPrefix string,
		validator string,
		rpc *tendermint.RPC,
		mutex *sync.Mutex,
		wg *sync.WaitGroup,
		allBalances map[string]map[string][]types.Amount,
	) {
		defer wg.Done()

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

		mutex.Lock()
		defer mutex.Unlock()

		if query != nil {
			queryInfos = append(queryInfos, query)
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

		allBalances[chainName][validator] = balances
	}

	for _, chain := range q.Config.Chains {
		allBalances[chain.Name] = map[string][]types.Amount{}
		for _, consumerChain := range chain.ConsumerChains {
			allBalances[consumerChain.Name] = map[string][]types.Amount{}
		}
	}

	for _, chain := range q.Config.Chains {
		rpc, _ := q.RPCs[chain.Name]

		for _, validator := range chain.Validators {
			wg.Add(1 + len(chain.ConsumerChains))

			go processChain(
				chain.Name,
				chain.BechWalletPrefix,
				validator.Address,
				rpc.RPC,
				&mutex,
				&wg,
				allBalances,
			)

			for consumerIndex, consumerChain := range chain.ConsumerChains {
				consumerRPC := rpc.Consumers[consumerIndex]

				go processChain(
					consumerChain.Name,
					consumerChain.BechWalletPrefix,
					validator.Address,
					consumerRPC,
					&mutex,
					&wg,
					allBalances,
				)
			}
		}
	}

	wg.Wait()

	return BalanceData{Balances: allBalances}, queryInfos
}

func (q *BalanceFetcher) Name() constants.FetcherName {
	return constants.FetcherNameBalance
}
