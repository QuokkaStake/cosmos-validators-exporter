package constants

type FetcherName string

type PriceFetcherName string

const (
	FetcherNameSlashingParams     FetcherName = "slashing-params"
	FetcherNameCommission         FetcherName = "commission"
	FetcherNameDelegations        FetcherName = "delegations"
	FetcherNameValidatorConsumers FetcherName = "validator-consumers"
	FetcherNameConsumerCommission FetcherName = "consumer-commission"

	FetcherNameUnbonds            FetcherName = "unbonds"
	FetcherNameSigningInfo        FetcherName = "signing-info"
	FetcherNameRewards            FetcherName = "rewards"
	FetcherNameBalance            FetcherName = "balance"
	FetcherNameSelfDelegation     FetcherName = "self-delegation"
	FetcherNameValidators         FetcherName = "validators"
	FetcherNameConsumerValidators FetcherName = "consumer-validators"
	FetcherNameConsumerInfo       FetcherName = "consumer-info"
	FetcherNameStakingParams      FetcherName = "staking_params"
	FetcherNamePrice              FetcherName = "price"
	FetcherNameNodeInfo           FetcherName = "node_info"
	FetcherNameInflation          FetcherName = "inflation"
	FetcherNameSupply             FetcherName = "supply"
	FetcherNameStub1              FetcherName = "stub1"
	FetcherNameStub2              FetcherName = "stub2"

	MetricsPrefix string = "cosmos_validators_exporter_"

	ValidatorStatusBonded = "BOND_STATUS_BONDED"

	HeaderBlockHeight = "Grpc-Metadata-X-Cosmos-Block-Height"

	CoingeckoBaseCurrency string = "usd"

	PriceFetcherNameCoingecko PriceFetcherName = "coingecko"
)
