package assets

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPanicOrFailPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	GetBytesOrPanic("not-existing")
}
