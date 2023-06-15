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

func eventTopic(block uint32, requestIndex uint16, eventIndex uint16) string {
	topic := fmt.Sprintf("fakeEvent:%d.%d.%d", block, requestIndex, eventIndex)
	return topic
}

func createEventLookupKeys(eventMap *collections.Map, contractID isc.Hname, maxBlocks uint32, maxRequests uint16, maxEvents uint16) {
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
	}
}

func validateEvents(t *testing.T, eventsInBytes [][]byte, maxRequests uint16, maxEvents uint16, blockFrom uint32, blockTo uint32) {
	require.Len(t, eventsInBytes, int(maxRequests*maxEvents)*int(blockTo-blockFrom+1))

	eventTopics := make([]string, 0)
	for _, eventBytes := range eventsInBytes {
		event, err := isc.NewEvent(eventBytes)
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

	eventMap := collections.NewMap(d, prefixRequestEvents)
	createEventLookupKeys(eventMap, contractID, maxBlocks, maxRequests, maxEventsPerRequest)

	events := getSmartContractEventsInternal(d, contractID, blockFrom, blockTo)
	validateEvents(t, events, maxRequests, maxEventsPerRequest, blockFrom, blockTo)
}
