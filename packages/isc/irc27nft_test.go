package isc_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestIRC27NFTSerialization(t *testing.T) {
	testMIME := "fakeMIME"
	testURL := "http://no.org"
	testName := "hi-name"
	metadata := isc.NewIRC27NFTMetadata(testMIME, testURL, testName, []interface{}{`{"trait_type": "Foo", "value": "Bar"}`})
	rwutil.BytesTest(t, metadata, isc.IRC27NFTMetadataFromBytes)
}

const sampleIRC27JSON = `{
"standard": "IRC27",
"version": "v1.0",
"name": "test-attr-2",
"type": "text/html; charset=UTF-8",
"uri": "https://google.de",
"attributes": [
	{
		"trait_type": "Base",
		"value": "Starfish"
	},
	{
		"trait_type": "Eyes",
		"value": "Big"
	}
]
}`

func TestIRC27FromJSON(t *testing.T) {
	parsed, err := isc.IRC27NFTMetadataFromBytes([]byte(sampleIRC27JSON))
	require.NoError(t, err)
	println(parsed)
}
