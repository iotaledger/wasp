package blocklog

import (
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

const EventLookupKeyLength = 8

// EventLookupKey is a globally unique reference to the event:
// block index + index of the request within block + index of the event within the request
type EventLookupKey [EventLookupKeyLength]byte

func NewEventLookupKey(blockIndex uint32, requestIndex, eventIndex uint16) *EventLookupKey {
	var ret EventLookupKey
	copy(ret[:4], codec.Uint32.Encode(blockIndex))
	copy(ret[4:6], codec.Uint16.Encode(requestIndex))
	copy(ret[6:8], codec.Uint16.Encode(eventIndex))
	return &ret
}

func (k *EventLookupKey) BlockIndex() uint32 {
	return codec.Uint32.MustDecode(k[:4])
}

func (k *EventLookupKey) RequestIndex() uint16 {
	return codec.Uint16.MustDecode(k[4:6])
}

func (k *EventLookupKey) RequestEventIndex() uint16 {
	return codec.Uint16.MustDecode(k[6:8])
}

func (k *EventLookupKey) Bytes() []byte {
	return bcs.MustMarshal(k)
}
