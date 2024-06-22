package utils

import (
	"main/pkg/constants"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestCompareStruct struct {
	value int
}

func TestCompareTwoBech32FirstInvalid(t *testing.T) {
	t.Parallel()

	_, err := CompareTwoBech32("test", "cosmos1xqz9pemz5e5zycaa89kys5aw6m8rhgsvtp9lt2")
	require.Error(t, err, "Error should be present!")
}

func TestCompareTwoBech32SecondInvalid(t *testing.T) {
	t.Parallel()

	_, err := CompareTwoBech32("cosmos1xqz9pemz5e5zycaa89kys5aw6m8rhgsvtp9lt2", "test")
	require.Error(t, err, "Error should be present!")
}

func TestCompareTwoBech32SecondEqual(t *testing.T) {
	t.Parallel()

	equal, err := CompareTwoBech32(
		"cosmos1xqz9pemz5e5zycaa89kys5aw6m8rhgsvtp9lt2",
		"cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
	)
	require.NoError(t, err, "Error should not be present!")
	assert.True(t, equal, "Bech addresses should be equal!")
}

func TestCompareTwoBech32SecondNotEqual(t *testing.T) {
	t.Parallel()

	equal, err := CompareTwoBech32(
		"cosmos1xqz9pemz5e5zycaa89kys5aw6m8rhgsvtp9lt2",
		"cosmos1c4k24jzduc365kywrsvf5ujz4ya6mwymy8vq4q",
	)
	require.NoError(t, err, "Error should not be present!")
	assert.False(t, equal, "Bech addresses should not be equal!")
}

func TestBoolToFloat64(t *testing.T) {
	t.Parallel()
	assert.InDelta(t, float64(1), BoolToFloat64(true), 0.001)
	assert.InDelta(t, float64(0), BoolToFloat64(false), 0.001)
}

func TestStrToFloat64(t *testing.T) {
	t.Parallel()
	assert.InDelta(t, 1.234, StrToFloat64("1.234"), 0.001)
}

func TestStrToFloat64Invalid(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	StrToFloat64("test")
}

func TestFilter(t *testing.T) {
	t.Parallel()

	array := []string{"true", "false"}
	filtered := Filter(array, func(s string) bool {
		return s == "true"
	})

	assert.Len(t, filtered, 1, "Array should have 1 entry!")
	assert.Equal(t, "true", filtered[0], "Value mismatch!")
}

func TestMap(t *testing.T) {
	t.Parallel()

	array := []int{2, 4}
	filtered := Map(array, func(v int) int {
		return v * 2
	})

	assert.Len(t, filtered, 2, "Array should have 2 entries!")
	assert.Equal(t, 4, filtered[0], "Value mismatch!")
	assert.Equal(t, 8, filtered[1], "Value mismatch!")
}

func TestFind(t *testing.T) {
	t.Parallel()

	array := []TestCompareStruct{{value: 2}, {value: 4}}
	value, found := Find(array, func(v TestCompareStruct) bool {
		return v.value == 2
	})

	assert.True(t, found)
	assert.NotNil(t, value)
	assert.Equal(t, 2, value.value)

	value2, found2 := Find(array, func(v TestCompareStruct) bool {
		return v.value == 3
	})

	assert.Nil(t, value2)
	assert.False(t, found2)
}

func TestFindIndex(t *testing.T) {
	t.Parallel()

	array := []TestCompareStruct{{value: 2}, {value: 4}}
	index, found := FindIndex(array, func(v TestCompareStruct) bool {
		return v.value == 4
	})

	assert.True(t, found)
	assert.Equal(t, 1, index)

	value2, found2 := FindIndex(array, func(v TestCompareStruct) bool {
		return v.value == 3
	})

	assert.Equal(t, 0, value2)
	assert.False(t, found2)
}

func TestChangeBech32Prefix(t *testing.T) {
	t.Parallel()

	_, err := ChangeBech32Prefix("test", "test")
	require.Error(t, err)

	value, err := ChangeBech32Prefix("cosmos1xqz9pemz5e5zycaa89kys5aw6m8rhgsvtp9lt2", "cosmosvaloper")
	require.NoError(t, err)
	require.Equal(t, "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e", value)
}

func TestGetBlockFromHeaderNoValue(t *testing.T) {
	t.Parallel()

	header := http.Header{}
	value, err := GetBlockHeightFromHeader(header)

	require.NoError(t, err)
	assert.Equal(t, int64(0), value)
}

func TestGetBlockFromHeaderInvalidValue(t *testing.T) {
	t.Parallel()

	header := http.Header{
		constants.HeaderBlockHeight: []string{"invalid"},
	}
	value, err := GetBlockHeightFromHeader(header)

	require.Error(t, err)
	assert.Equal(t, int64(0), value)
}

func TestGetBlockFromHeaderValidValue(t *testing.T) {
	t.Parallel()

	header := http.Header{
		constants.HeaderBlockHeight: []string{"123"},
	}
	value, err := GetBlockHeightFromHeader(header)

	require.NoError(t, err)
	assert.Equal(t, int64(123), value)
}
