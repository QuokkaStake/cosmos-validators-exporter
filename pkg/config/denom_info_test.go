package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestDenomInfoValidateNoDenom(t *testing.T) {
	t.Parallel()

	denom := DenomInfo{}
	err := denom.Validate()
	require.Error(t, err)
}

func TestDenomInfoValidateNoDisplayDenom(t *testing.T) {
	t.Parallel()

	denom := DenomInfo{Denom: "ustake"}
	err := denom.Validate()
	require.Error(t, err)
}

func TestDenomInfoValidateValid(t *testing.T) {
	t.Parallel()

	denom := DenomInfo{Denom: "ustake", DisplayDenom: "stake"}
	err := denom.Validate()
	require.NoError(t, err)
}

func TestDenomInfoDisplayWarningNoCoingecko(t *testing.T) {
	t.Parallel()

	denom := DenomInfo{Denom: "ustake", DisplayDenom: "stake"}
	warnings := denom.DisplayWarnings(&Chain{Name: "test"})
	assert.NotEmpty(t, warnings)
}

func TestDenomInfoDisplayWarningEmpty(t *testing.T) {
	t.Parallel()

	denom := DenomInfo{Denom: "ustake", DisplayDenom: "stake", CoingeckoCurrency: "test"}
	warnings := denom.DisplayWarnings(&Chain{Name: "test"})
	assert.Empty(t, warnings)
}

func TestDenomInfosFind(t *testing.T) {
	t.Parallel()

	denoms := DenomInfos{{Denom: "ustake"}}
	assert.NotNil(t, denoms.Find("ustake"))
	assert.Nil(t, denoms.Find("uatom"))
}
