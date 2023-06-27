package isc_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestIRC27NFTSerialization(t *testing.T) {
	testMIME := "fakeMIME"
	testURL := "http://no.org"
	testName := "hi-name"
	metadata := isc.NewIRC27NFTMetadata(testMIME, testURL, testName)
	rwutil.BytesTest(t, metadata, isc.IRC27NFTMetadataFromBytes)
}
