package constants

type FetcherName string

const (
	FetcherNameSlashingParams      FetcherName = "slashing-params"
	FetcherNameSoftOptOutThreshold FetcherName = "soft-opt-out-threshold"
	FetcherNameCommission          FetcherName = "commission"
	FetcherNameDelegations         FetcherName = "delegations"
	FetcherNameUnbonds             FetcherName = "unbonds"

	MetricsPrefix string = "cosmos_validators_exporter_"
)
