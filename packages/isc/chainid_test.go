package isc

import (
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestChainIDSerialization(t *testing.T) {
	chainID := ChainID(*iotatest.RandomAddress())
	bcs.TestCodec(t, chainID)
	rwutil.BytesTest(t, chainID, ChainIDFromBytes)
	rwutil.StringTest(t, chainID, ChainIDFromString)
}
