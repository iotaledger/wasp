package blocklog

import (
	"fmt"
	"testing"
	"time"

	"github.com/samber/lo"
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
		pruneRequestLookupTable(d, requestIDDigest0, blockToPrune)
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

	registry := collections.NewArray(d, PrefixBlockRegistry)

	eventMap := collections.NewMap(d, prefixRequestEvents)
	createEventLookupKeys(registry, eventMap, contractID, maxBlocks, maxRequests, maxEventsPerRequest)

	events := getSmartContractEventsInternal(d, contractID, blockFrom, blockTo)
	validateEvents(t, events, maxRequests, maxEventsPerRequest, blockFrom, blockTo)
}
