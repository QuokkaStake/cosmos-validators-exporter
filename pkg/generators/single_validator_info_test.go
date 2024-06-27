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

func TestSingleValidatorInfoGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewSingleValidatorInfoGenerator([]*config.Chain{}, loggerPkg.GetNopLogger())
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestSingleValidatorInfoGeneratorNoChainValidators(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{Name: "chain"}}
	state := statePkg.NewState()
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{})
	generator := NewSingleValidatorInfoGenerator(chains, loggerPkg.GetNopLogger())
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 0, testutil.CollectAndCount(gauge))
}

func TestSingleValidatorInfoGeneratorNotFound(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:       "chain",
		Validators: []config.Validator{{Address: "validator"}},
	}}
	state := statePkg.NewState()
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
	generator := NewSingleValidatorInfoGenerator(chains, loggerPkg.GetNopLogger())
	results := generator.Generate(state)
	assert.Len(t, results, 5)

	for _, metric := range results {
		gauge, ok := metric.(*prometheus.GaugeVec)
		assert.True(t, ok)
		assert.Equal(t, 0, testutil.CollectAndCount(gauge))
	}
}

func TestSingleValidatorInfoGeneratorActive(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:       "chain",
		BaseDenom:  "uatom",
		Denoms:     config.DenomInfos{{Denom: "uatom", DisplayDenom: "atom", DenomExponent: 6}},
		Validators: []config.Validator{{Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"}},
	}}
	state := statePkg.NewState()
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{
		Validators: map[string]*types.ValidatorsResponse{
			"chain": {
				Validators: []types.Validator{
					{
						DelegatorShares: math.LegacyMustNewDecFromStr("3000000"),
						OperatorAddress: "cosmosvaloper1c4k24jzduc365kywrsvf5ujz4ya6mwympnc4en",
						Status:          constants.ValidatorStatusBonded,
					},
					{
						DelegatorShares: math.LegacyMustNewDecFromStr("2000000"),
						OperatorAddress: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
						Status:          constants.ValidatorStatusBonded,
						Description: types.ValidatorDescription{
							Moniker:         "moniker",
							SecurityContact: "contact",
							Website:         "website",
							Details:         "details",
							Identity:        "identity",
						},
						Commission: types.ValidatorCommission{
							CommissionRates: types.ValidatorCommissionRates{
								Rate:          math.LegacyMustNewDecFromStr("0.05"),
								MaxRate:       math.LegacyMustNewDecFromStr("0.2"),
								MaxChangeRate: math.LegacyMustNewDecFromStr("0.01"),
							},
						},
					},
					{
						DelegatorShares: math.LegacyMustNewDecFromStr("1000000"),
						OperatorAddress: "cosmosvaloper14lultfckehtszvzw4ehu0apvsr77afvyju5zzy",
						Status:          constants.ValidatorStatusBonded,
					},
				},
			},
		},
	})
	generator := NewSingleValidatorInfoGenerator(chains, loggerPkg.GetNopLogger())
	results := generator.Generate(state)
	assert.Len(t, results, 5)

	validatorInfoGauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, float64(1), testutil.ToFloat64(validatorInfoGauge.With(prometheus.Labels{
		"chain":            "chain",
		"address":          "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
		"moniker":          "moniker",
		"details":          "details",
		"identity":         "identity",
		"security_contact": "contact",
		"website":          "website",
	})), 0.01)

	isJailed, ok := results[1].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Zero(t, testutil.ToFloat64(isJailed.With(prometheus.Labels{
		"chain":   "chain",
		"address": "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
	})))

	commissionMaxGauge, ok := results[2].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, 0.2, testutil.ToFloat64(commissionMaxGauge.With(prometheus.Labels{
		"chain":   "chain",
		"address": "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
	})), 0.01)

	commissionMaxChangeGauge, ok := results[3].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, 0.01, testutil.ToFloat64(commissionMaxChangeGauge.With(prometheus.Labels{
		"chain":   "chain",
		"address": "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
	})), 0.01)

	delegationsGauge, ok := results[4].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, float64(2), testutil.ToFloat64(delegationsGauge.With(prometheus.Labels{
		"chain":   "chain",
		"address": "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
		"denom":   "atom",
	})), 0.01)
}
