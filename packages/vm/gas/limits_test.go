package gas_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

func TestGasLimitsSerialization(t *testing.T) {
	limits0 := &gas.Limits{
		MaxGasPerBlock:         456,
		MinGasPerRequest:       123,
		MaxGasPerRequest:       789,
		MaxGasExternalViewCall: 12342,
	}

	b := limits0.Bytes()
	limits1, err := gas.LimitsFromBytes(b)
	require.NoError(t, err)
	require.Equal(t, limits0, limits1)
}
