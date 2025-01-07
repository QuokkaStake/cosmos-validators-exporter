package state

import (
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStateGetNoValue(t *testing.T) {
	t.Parallel()

	state := NewState()
	value, found := state.Get(constants.FetcherNameCommission)
	require.False(t, found)
	require.Nil(t, value)
}

func TestStateConvertNoValue(t *testing.T) {
	t.Parallel()

	value, found := Convert[fetchersPkg.CommissionData](nil)
	require.False(t, found)
	require.Zero(t, value)
}

func TestStateConvertWrongType(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	Convert[int64]("string")
}

func TestStateConvertNilPointer(t *testing.T) {
	t.Parallel()

	var data *fetchersPkg.CommissionData

	value, found := Convert[*fetchersPkg.CommissionData](data)
	require.False(t, found)
	require.Zero(t, value)
}

func TestStateConvertOk(t *testing.T) {
	t.Parallel()

	value, found := Convert[string]("string")
	require.True(t, found)
	require.Equal(t, "string", value)
}
