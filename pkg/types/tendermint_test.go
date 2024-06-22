package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatorActive(t *testing.T) {
	t.Parallel()

	assert.False(t, Validator{Status: "BOND_STATUS_UNBONDED"}.Active())
	assert.True(t, Validator{Status: "BOND_STATUS_BONDED"}.Active())
}

func TestResponseAmountToAmount(t *testing.T) {
	t.Parallel()

	responseAmount := ResponseAmount{
		Amount: "1.23",
		Denom:  "ustake",
	}

	converted := responseAmount.ToAmount()

	assert.InDelta(t, 1.23, 0.0001, converted.Amount)
	assert.Equal(t, "ustake", converted.Denom)
}
