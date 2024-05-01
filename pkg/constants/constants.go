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
	FetcherNameStakingParams       FetcherName = "staking_params"

	MetricsPrefix string = "cosmos_validators_exporter_"
)
