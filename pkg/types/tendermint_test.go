package types

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestValidatorActive(t *testing.T) {
	t.Parallel()

	assert.False(t, Validator{Status: "BOND_STATUS_UNBONDED"}.Active())
	assert.True(t, Validator{Status: "BOND_STATUS_BONDED"}.Active())
}

func TestResponseAmountToAmount(t *testing.T) {
	t.Parallel()

	amount, err := math.LegacyNewDecFromStr("1.23")
	require.NoError(t, err)

	responseAmount := ResponseAmount{
		Amount: amount,
		Denom:  "ustake",
	}

	converted := responseAmount.ToAmount()

	assert.InDelta(t, 1.23, 0.0001, converted.Amount)
	assert.Equal(t, "ustake", converted.Denom)
}
