package generators

import (
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/fetchers"
	loggerPkg "main/pkg/logger"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"cosmossdk.io/math"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"
)

func TestValidatorActiveGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.State{}
	generator := NewValidatorActiveGenerator([]*config.Chain{}, loggerPkg.GetNopLogger())
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestValidatorActiveGeneratorNoConsumerValidators(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{Name: "chain"}}
	state := statePkg.State{}
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{})
	generator := NewValidatorActiveGenerator(chains, loggerPkg.GetNopLogger())
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestValidatorActiveGeneratorNoChainValidators(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{Name: "chain"}}
	state := statePkg.State{}
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{})
	state.Set(constants.FetcherNameConsumerValidators, fetchers.ConsumerValidatorsData{})
	generator := NewValidatorActiveGenerator(chains, loggerPkg.GetNopLogger())
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 0, testutil.CollectAndCount(gauge))
}

func TestValidatorActiveGeneratorNotFound(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:       "chain",
		Validators: []config.Validator{{Address: "validator"}},
	}}
	state := statePkg.State{}
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{
		Validators: map[string]*types.ValidatorsResponse{
			"chain": {
				Validators: []types.Validator{
					{DelegatorShares: math.LegacyMustNewDecFromStr("2"), OperatorAddress: "first"},
					{DelegatorShares: math.LegacyMustNewDecFromStr("1"), OperatorAddress: "second"},
					{DelegatorShares: math.LegacyMustNewDecFromStr("3"), OperatorAddress: "third"},
				},
			},
		},
	})
	state.Set(constants.FetcherNameConsumerValidators, fetchers.ConsumerValidatorsData{})
	generator := NewValidatorActiveGenerator(chains, loggerPkg.GetNopLogger())
	results := generator.Generate(state)
	assert.Len(t, results, 1)

	isActive, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 0, testutil.CollectAndCount(isActive))
}

func TestValidatorActiveGeneratorProvider(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:       "chain",
		BaseDenom:  "uatom",
		Denoms:     config.DenomInfos{{Denom: "uatom", DisplayDenom: "atom", DenomExponent: 6}},
		Validators: []config.Validator{{Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"}},
	}}
	state := statePkg.State{}
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{
		Validators: map[string]*types.ValidatorsResponse{
			"chain": {
				Validators: []types.Validator{
					{
						OperatorAddress: "cosmosvaloper1c4k24jzduc365kywrsvf5ujz4ya6mwympnc4en",
					},
					{
						OperatorAddress: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
						Status:          constants.ValidatorStatusBonded,
					},
					{
						OperatorAddress: "cosmosvaloper14lultfckehtszvzw4ehu0apvsr77afvyju5zzy",
					},
				},
			},
		},
	})
	state.Set(constants.FetcherNameConsumerValidators, fetchers.ConsumerValidatorsData{})
	generator := NewValidatorActiveGenerator(chains, loggerPkg.GetNopLogger())
	results := generator.Generate(state)
	assert.Len(t, results, 1)

	isActive, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, 1, testutil.ToFloat64(isActive.With(prometheus.Labels{
		"chain":   "chain",
		"address": "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
	})), 0.01)
}

func TestValidatorActiveGeneratorConsumer(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:      "chain",
		BaseDenom: "uatom",
		Denoms:    config.DenomInfos{{Denom: "uatom", DisplayDenom: "atom", DenomExponent: 6}},
		Validators: []config.Validator{{
			Address:          "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
			ConsensusAddress: "cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		}, {Address: "otheraddress"}},
		ConsumerChains: []*config.ConsumerChain{
			{Name: "consumer"},
			{Name: "otherconsumer"},
		},
	}}
	state := statePkg.State{}
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{
		Validators: map[string]*types.ValidatorsResponse{
			"chain": {
				Validators: []types.Validator{
					{
						OperatorAddress: "cosmosvaloper1c4k24jzduc365kywrsvf5ujz4ya6mwympnc4en",
					},
					{
						OperatorAddress: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
						Status:          constants.ValidatorStatusBonded,
					},
					{
						OperatorAddress: "cosmosvaloper14lultfckehtszvzw4ehu0apvsr77afvyju5zzy",
					},
				},
			},
		},
	})
	state.Set(constants.FetcherNameConsumerValidators, fetchers.ConsumerValidatorsData{
		Validators: map[string]*types.ConsumerValidatorsResponse{
			"consumer": {
				Validators: []types.ConsumerValidator{
					{ProviderAddress: "cosmosvalcons1rnknqueazcju3x4knpnhvpksudanva4f08558z"},
					{ProviderAddress: "cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc"},
				},
			},
		},
	})
	generator := NewValidatorActiveGenerator(chains, loggerPkg.GetNopLogger())
	results := generator.Generate(state)
	assert.Len(t, results, 1)

	isActive, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, 1, testutil.ToFloat64(isActive.With(prometheus.Labels{
		"chain":   "chain",
		"address": "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
	})), 0.01)
}
