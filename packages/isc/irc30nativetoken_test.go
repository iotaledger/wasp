package isc_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestIRC30NativeTokenSerialization(t *testing.T) {
	testName := "TestyTest"
	testSymbol := "TT"
	testDecimals := uint8(8)

	metadata := isc.NewIRC30NativeTokenMetadata(testName, testSymbol, testDecimals)
	rwutil.BytesTest(t, metadata, isc.IRC30NativeTokenMetadataFromBytes)
}
