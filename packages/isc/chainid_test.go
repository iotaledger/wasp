package isc

import (
	"testing"

	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/sui-go/sui/suitest"
)

func TestChainIDSerialization(t *testing.T) {
	chainID := ChainID(*suitest.RandomAddress())
	bcs.TestCodec(t, chainID)
	rwutil.BytesTest(t, chainID, ChainIDFromBytes)
	rwutil.StringTest(t, chainID, ChainIDFromString)
}
