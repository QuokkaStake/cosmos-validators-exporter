package constants

type FetcherName string

const (
	FetcherNameSlashingParams      FetcherName = "slashing-params"
	FetcherNameSoftOptOutThreshold FetcherName = "soft-opt-out-threshold"
	FetcherNameCommission          FetcherName = "commission"
	FetcherNameDelegations         FetcherName = "delegations"
	FetcherNameUnbonds             FetcherName = "unbonds"
	FetcherNameSigningInfo         FetcherName = "signing-info"
	FetcherNameRewards             FetcherName = "rewards"
	FetcherNameBalance             FetcherName = "balance"
	FetcherNameSelfDelegation      FetcherName = "self-delegation"
	FetcherNameValidators          FetcherName = "validators"
	FetcherNameConsumerValidators  FetcherName = "consumer-validators"
	FetcherNameStakingParams       FetcherName = "staking_params"
	FetcherNamePrice               FetcherName = "price"
	FetcherNameNodeInfo            FetcherName = "node_info"

	MetricsPrefix string = "cosmos_validators_exporter_"

	ValidatorStatusBonded = "BOND_STATUS_BONDED"

	HeaderBlockHeight = "Grpc-Metadata-X-Cosmos-Block-Height"
)
