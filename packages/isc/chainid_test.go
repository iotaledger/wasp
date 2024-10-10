package isc

import (
	"testing"

	"github.com/iotaledger/wasp/clients/iota-go/iotago/suitest"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestChainIDSerialization(t *testing.T) {
	chainID := ChainID(*suitest.RandomAddress())
	bcs.TestCodec(t, chainID)
	rwutil.BytesTest(t, chainID, ChainIDFromBytes)
	rwutil.StringTest(t, chainID, ChainIDFromString)
}
