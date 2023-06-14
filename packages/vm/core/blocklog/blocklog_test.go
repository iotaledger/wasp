package blocklog

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func TestSerdeRequestReceipt(t *testing.T) {
	nonce := uint64(time.Now().UnixNano())
	req := isc.NewOffLedgerRequest(isc.RandomChainID(), isc.Hn("0"), isc.Hn("0"), nil, nonce, gas.LimitsDefault.MaxGasPerRequest)
	signedReq := req.Sign(cryptolib.NewKeyPair())
	rec := &RequestReceipt{
		Request: signedReq,
	}
	forward := rec.Bytes()
	back, err := RequestReceiptFromBytes(forward)
	require.NoError(t, err)
	require.EqualValues(t, forward, back.Bytes())
}

func createRequestLookupKeys(blocks uint32) []byte {
	keys := make(RequestLookupKeyList, 0)

	for blockIndex := uint32(0); blockIndex < blocks; blockIndex++ {
		for reqIndex := uint16(0); reqIndex < 3; reqIndex++ {
			key := NewRequestLookupKey(blockIndex, reqIndex)

			keys = append(keys, key)
		}
	}

	return keys.Bytes()
}

func validatePrunedRequestIndexLookupBlock(t *testing.T, partition kv.KVStore, contract kv.Key, prunedBlockIndex uint32) {
	requestLookup := collections.NewMap(partition, prefixRequestLookupIndex)
	requestKeys, err := RequestLookupKeyListFromBytes(requestLookup.GetAt([]byte(contract)))
	require.NoError(t, err)

	for _, requestKey := range requestKeys {
		require.True(t, requestKey.BlockIndex() != prunedBlockIndex)
	}
}

func TestPruneRequestIndexLookupTable(t *testing.T) {
	const maxBlocks = 42
	const blockToPrune = 33

	requestIDDigest0 := kv.Key("0")
	requestIDDigest1 := kv.Key("1")

	d := dict.Dict{}

	requestIndexLUT := collections.NewMap(d, prefixRequestLookupIndex)
	requestIndexLUT.SetAt([]byte(requestIDDigest0), createRequestLookupKeys(maxBlocks))
	requestIndexLUT.SetAt([]byte(requestIDDigest1), createRequestLookupKeys(maxBlocks))

	require.NotPanics(t, func() {
		pruneRequestLookupByBlockIndex(d, blockToPrune)
	})

	validatePrunedRequestIndexLookupBlock(t, d, requestIDDigest0, blockToPrune)
	validatePrunedRequestIndexLookupBlock(t, d, requestIDDigest1, blockToPrune)
}
