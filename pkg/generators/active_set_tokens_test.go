package generators

import (
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/fetchers"
	"main/pkg/types"
	"testing"

	statePkg "main/pkg/state"

	"github.com/guregu/null/v5"

	"cosmossdk.io/math"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"
)

func TestActiveSetTokensGeneratorNoValidators(t *testing.T) {
	t.Parallel()

	state := statePkg.State{}
	generator := NewActiveSetTokensGenerator([]*config.Chain{})
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestActiveSetTokensGeneratorNoStakingParams(t *testing.T) {
	t.Parallel()

	state := statePkg.State{}
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{})
	generator := NewActiveSetTokensGenerator([]*config.Chain{})
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestActiveSetTokensGeneratorNoChainValidators(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{Name: "chain"}}
	state := statePkg.State{}
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{})
	state.Set(constants.FetcherNameStakingParams, fetchers.StakingParamsData{})
	generator := NewActiveSetTokensGenerator(chains)
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 0, testutil.CollectAndCount(gauge))
}

func TestActiveSetTokensGeneratorNoChainStakingParams(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{Name: "chain"}}
	state := statePkg.State{}
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{
		Validators: map[string]*types.ValidatorsResponse{
			"chain": {},
		},
	})
	state.Set(constants.FetcherNameStakingParams, fetchers.StakingParamsData{})
	generator := NewActiveSetTokensGenerator(chains)
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 0, testutil.CollectAndCount(gauge))
}

func TestActiveSetTokensGeneratorNotEnoughValidators(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:      "chain",
		BaseDenom: "uatom",
		Denoms:    config.DenomInfos{{Denom: "uatom", DisplayDenom: "atom", DenomExponent: 6}},
	}}
	state := statePkg.State{}
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{
		Validators: map[string]*types.ValidatorsResponse{
			"chain": {
				Validators: []types.Validator{
					{DelegatorShares: math.LegacyMustNewDecFromStr("2"), Status: constants.ValidatorStatusBonded},
					{DelegatorShares: math.LegacyMustNewDecFromStr("1"), Status: constants.ValidatorStatusBonded},
					{DelegatorShares: math.LegacyMustNewDecFromStr("3"), Status: constants.ValidatorStatusBonded},
				},
			},
		},
	})
	state.Set(constants.FetcherNameStakingParams, fetchers.StakingParamsData{
		Params: map[string]*types.StakingParamsResponse{
			"chain": {
				StakingParams: types.StakingParams{MaxValidators: 100},
			},
		},
	})
	generator := NewActiveSetTokensGenerator(chains)
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Zero(t, testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain": "chain",
		"denom": "atom",
	})))
}

func TestActiveSetTokensGeneratorEnoughValidators(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:      "chain",
		BaseDenom: "uatom",
		Denoms:    config.DenomInfos{{Denom: "uatom", DisplayDenom: "atom", DenomExponent: 6}},
	}}
	state := statePkg.State{}
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{
		Validators: map[string]*types.ValidatorsResponse{
			"chain": {
				Validators: []types.Validator{
					{DelegatorShares: math.LegacyMustNewDecFromStr("2000000"), Status: constants.ValidatorStatusBonded},
					{DelegatorShares: math.LegacyMustNewDecFromStr("1000000")},
					{DelegatorShares: math.LegacyMustNewDecFromStr("3000000"), Status: constants.ValidatorStatusBonded},
				},
			},
		},
	})
	state.Set(constants.FetcherNameStakingParams, fetchers.StakingParamsData{
		Params: map[string]*types.StakingParamsResponse{
			"chain": {
				StakingParams: types.StakingParams{MaxValidators: 2},
			},
		},
	})
	generator := NewActiveSetTokensGenerator(chains)
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, float64(2), testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain": "chain",
		"denom": "atom",
	})), 0.01)
}

func TestActiveSetTokensGeneratorDenomIgnored(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:      "chain",
		BaseDenom: "uatom",
		Denoms:    config.DenomInfos{{Denom: "uatom", Ignore: null.BoolFrom(true)}},
	}}
	state := statePkg.State{}
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{
		Validators: map[string]*types.ValidatorsResponse{
			"chain": {
				Validators: []types.Validator{
					{DelegatorShares: math.LegacyMustNewDecFromStr("2000000"), Status: constants.ValidatorStatusBonded},
					{DelegatorShares: math.LegacyMustNewDecFromStr("1000000")},
					{DelegatorShares: math.LegacyMustNewDecFromStr("3000000"), Status: constants.ValidatorStatusBonded},
				},
			},
		},
	})
	state.Set(constants.FetcherNameStakingParams, fetchers.StakingParamsData{
		Params: map[string]*types.StakingParamsResponse{
			"chain": {
				StakingParams: types.StakingParams{MaxValidators: 2},
			},
		},
	})
	generator := NewActiveSetTokensGenerator(chains)
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Zero(t, testutil.CollectAndCount(gauge))
}
