package isc

import (
	"testing"

	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestChainIDSerialization(t *testing.T) {
	chainID := RandomChainID()
	bcs.TestCodec(t, chainID)
	rwutil.BytesTest(t, chainID, ChainIDFromBytes)
	rwutil.StringTest(t, chainID, ChainIDFromString)
}
