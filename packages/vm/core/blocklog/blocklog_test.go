package blocklog

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/isctest"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/collections"
	"github.com/iotaledger/wasp/v2/packages/kv/dict"
	"github.com/iotaledger/wasp/v2/packages/parameters/parameterstest"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

func TestSerdeRequestReceipt(t *testing.T) {
	nonce := uint64(time.Now().UnixNano())
	req := isc.NewOffLedgerRequest(isctest.RandomChainID(), isc.NewMessage(isc.Hn("0"), isc.Hn("0")), nonce, gas.LimitsDefault.MaxGasPerRequest)
	signedReq := req.Sign(cryptolib.NewKeyPair())
	rec := &RequestReceipt{
		Request: signedReq,
	}
	forward := rec.Bytes()
	back, err := RequestReceiptFromBytes(forward, rec.BlockIndex, rec.RequestIndex)
	require.NoError(t, err)
	require.EqualValues(t, forward, back.Bytes())
}

func createRequestLookupKeys(blocks uint32, requests uint16) []byte {
	keys := make(RequestLookupKeyList, 0)

	for blockIndex := uint32(0); blockIndex < blocks; blockIndex++ {
		for reqIndex := uint16(0); reqIndex < requests; reqIndex++ {
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
	const requestsToCreate = 4
	const blockToPrune = 33

	requestIDDigest0 := [8]byte{0, 0, 0, 0, 0, 0, 0, 0}
	requestIDDigest1 := [8]byte{0, 0, 0, 0, 0, 0, 0, 1}

	d := dict.Dict{}

	requestIndexLUT := collections.NewMap(d, prefixRequestLookupIndex)
	requestIndexLUT.SetAt(requestIDDigest0[:], createRequestLookupKeys(maxBlocks, requestsToCreate))
	requestIndexLUT.SetAt(requestIDDigest1[:], createRequestLookupKeys(maxBlocks, requestsToCreate))

	require.NotPanics(t, func() {
		NewStateWriter(d).pruneRequestLookupTable(requestIDDigest0, blockToPrune)
	})

	validatePrunedRequestIndexLookupBlock(t, d, kv.Key(requestIDDigest0[:]), blockToPrune)
	digest0Size := len(requestIndexLUT.GetAt(requestIDDigest0[:]))
	digest1Size := len(requestIndexLUT.GetAt(requestIDDigest1[:]))

	// Four requests in total should be removed in digest0, therefore the amount of bytes removed should be len(RequestLookupKey)*requestsToCreate
	require.Equal(t, len(RequestLookupKey{})*requestsToCreate, digest1Size-digest0Size)
}

func TestBlockInfoMarshalling(t *testing.T) {
	t.Run("v0", func(t *testing.T) {
		const v0hex = "002a00000000b421501f01000000640000000000000000000000000000000100000000000000000000000000000001000000000000000000000000000000e80300000000000000000000000000009edb91da930100000000000000000000005c2605000000000000000000000000000000000000000000000000000000000000000000000000000000000000000204696f746104494f5441000004496f746104494f544104494f54410f687474703a2f2f696f74612e6f726709e0afdabfb6f592bd8a010a0008000200e807f403"
		var blockIndex uint32 = 42
		expected := &BlockInfo{
			SchemaVersion:         blockInfoSchemaVersion0,
			BlockIndex:            blockIndex,
			Timestamp:             time.Unix(1234, 0),
			PreviousAnchor:        nil,
			L1Params:              parameterstest.L1Mock,
			TotalRequests:         10,
			NumSuccessfulRequests: 8,
			NumOffLedgerRequests:  2,
			GasBurned:             1000,
			GasFeeCharged:         500,
			Entropy:               hashing.HashData(bcs.MustMarshal(&blockIndex)),
		}

		bi, err := BlockInfoFromBytes(lo.Must(hex.DecodeString(v0hex)))
		require.NoError(t, err)
		require.EqualValues(t, expected, bi)
	})

	t.Run("v6", func(t *testing.T) {
		biv1 := &BlockInfo{
			SchemaVersion:         blockInfoSchemaVersionAddedEntropy,
			BlockIndex:            42,
			Timestamp:             time.Unix(1234, 0),
			PreviousAnchor:        nil,
			L1Params:              parameterstest.L1Mock,
			TotalRequests:         10,
			NumSuccessfulRequests: 8,
			NumOffLedgerRequests:  2,

			GasBurned:     1000,
			GasFeeCharged: 500,
			Entropy:       hashing.PseudoRandomHash(nil),
		}

		bi, err := BlockInfoFromBytes(biv1.Bytes())
		require.NoError(t, err)
		require.EqualValues(t, biv1, bi)
	})

	t.Run("v7", func(t *testing.T) {
		biv1 := &BlockInfo{
			SchemaVersion:         blockInfoSchemaVersionAddedGasCoinTopUp,
			BlockIndex:            42,
			Timestamp:             time.Unix(1234, 0),
			PreviousAnchor:        nil,
			L1Params:              parameterstest.L1Mock,
			TotalRequests:         10,
			NumSuccessfulRequests: 8,
			NumOffLedgerRequests:  2,

			GasBurned:     1000,
			GasFeeCharged: 500,
			Entropy:       hashing.PseudoRandomHash(nil),
			GasCoinTopUp:  1234,
		}

		bi, err := BlockInfoFromBytes(biv1.Bytes())
		require.NoError(t, err)
		require.EqualValues(t, biv1, bi)
	})
}
