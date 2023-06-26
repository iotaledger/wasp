package isc

import (
	"testing"

	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestChainIDSerialization(t *testing.T) {
	chainID := RandomChainID()
	rwutil.ReadWriteTest(t, &chainID, new(ChainID))
	rwutil.BytesTest(t, chainID, ChainIDFromBytes)
	rwutil.StringTest(t, chainID, ChainIDFromString)
}
