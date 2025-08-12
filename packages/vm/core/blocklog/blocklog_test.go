package blocklog

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/collections"
	"github.com/iotaledger/wasp/v2/packages/kv/dict"
	"github.com/iotaledger/wasp/v2/packages/parameters/parameterstest"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

func TestSerdeRequestReceipt(t *testing.T) {
	nonce := uint64(time.Now().UnixNano())
	req := isc.NewOffLedgerRequest(isc.NewMessage(isc.Hn("0"), isc.Hn("0")), nonce, gas.LimitsDefault.MaxGasPerRequest)
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

func eventTopic(block uint32, requestIndex uint16, eventIndex uint16) string {
	topic := fmt.Sprintf("fakeEvent:%d.%d.%d", block, requestIndex, eventIndex)
	return topic
}

func createEventLookupKeys(registryArray *collections.Array, eventMap *collections.Map, contractID isc.Hname, maxBlocks uint32, maxRequests uint16, maxEvents uint16) {
	for blockIndex := uint32(0); blockIndex < maxBlocks; blockIndex++ {
		for reqIndex := uint16(0); reqIndex < maxRequests; reqIndex++ {
			for eventIndex := uint16(0); eventIndex < maxEvents; eventIndex++ {
				key := NewEventLookupKey(blockIndex, reqIndex, eventIndex).Bytes()
				topic := eventTopic(blockIndex, reqIndex, eventIndex)

				event := isc.Event{
					Topic:      topic,
					Payload:    nil,
					Timestamp:  uint64(time.Now().UnixNano()),
					ContractID: contractID,
				}

				eventMap.SetAt(key, event.Bytes())
			}
		}

		registryArray.Push([]byte{0})
	}
}

func validateEvents(t *testing.T, eventsInBytes [][]byte, maxRequests uint16, maxEvents uint16, blockFrom uint32, blockTo uint32) {
	require.Len(t, eventsInBytes, int(maxRequests*maxEvents)*int(blockTo-blockFrom+1))

	eventTopics := make([]string, 0)
	for _, eventBytes := range eventsInBytes {
		event, err := isc.EventFromBytes(eventBytes)
		require.NoError(t, err)
		eventTopics = append(eventTopics, event.Topic)
	}

	for blockIndex := blockFrom; blockIndex <= blockTo; blockIndex++ {
		for reqIndex := uint16(0); reqIndex < maxRequests; reqIndex++ {
			for eventIndex := uint16(0); eventIndex < maxEvents; eventIndex++ {
				topic := eventTopic(blockIndex, reqIndex, eventIndex)
				require.True(t, lo.Contains(eventTopics, topic))
			}
		}
	}
}

func TestGetEventsInternal(t *testing.T) {
	const maxBlocks = 20
	const maxRequests = 5
	const maxEventsPerRequest = 10

	const blockFrom = 5
	const blockTo = blockFrom + 5

	contractID := isc.Hn("testytest")

	d := dict.Dict{}

	registry := collections.NewArray(d, prefixBlockRegistry)

	eventMap := collections.NewMap(d, prefixRequestEvents)
	createEventLookupKeys(registry, eventMap, contractID, maxBlocks, maxRequests, maxEventsPerRequest)

	events := NewStateWriter(d).getSmartContractEventsInternal(EventsForContractQuery{contractID, &BlockRange{blockFrom, blockTo}})
	validateEvents(t, events, maxRequests, maxEventsPerRequest, blockFrom, blockTo)
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
