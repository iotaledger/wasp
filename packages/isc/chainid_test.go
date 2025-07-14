package isc_test

import (
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/isctest"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestChainIDSerialization(t *testing.T) {
	chainID := isc.ChainID(*iotatest.RandomAddress())
	bcs.TestCodec(t, chainID)
	rwutil.BytesTest(t, chainID, isc.ChainIDFromBytes)
	rwutil.StringTest(t, chainID, isc.ChainIDFromString)

	bcs.TestCodecAndHash(t, isctest.TestChainID, "b4ff315a20ce")
}
