package isc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChainID(t *testing.T) {
	chainID := RandomChainID()
	chainIDStr := chainID.String()

	chainIDFromBytes, err := ChainIDFromBytes(chainID.Bytes())
	assert.NoError(t, err)
	assert.EqualValues(t, chainIDFromBytes, chainID)

	chainIDFromString, err := ChainIDFromString(chainIDStr)
	assert.NoError(t, err)
	assert.EqualValues(t, chainIDFromString, chainID)
}
