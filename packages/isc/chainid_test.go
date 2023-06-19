package isc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChainID(t *testing.T) {
	chainID := RandomChainID()
	chainIDStr := chainID.String()

	chainIDFromBytes, err := ChainIDFromBytes(chainID.Bytes())
	require.NoError(t, err)
	require.EqualValues(t, chainIDFromBytes, chainID)

	chainIDFromString, err := ChainIDFromString(chainIDStr)
	require.NoError(t, err)
	require.EqualValues(t, chainIDFromString, chainID)
}
