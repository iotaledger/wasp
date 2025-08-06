package isc_test

import (
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/isctest"
	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
)

func TestChainIDSerialization(t *testing.T) {
	chainID := isc.ChainID(*iotatest.RandomAddress())
	bcs.TestCodec(t)
	rwutil.BytesTest(t, isc.ChainIDFromBytes)
	rwutil.StringTest(t, isc.ChainIDFromString)

	bcs.TestCodecAndHash(t, isctest.TestChainID, "b4ff315a20ce")
}
