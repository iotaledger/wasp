package blocklog

import (
	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
)

const EventLookupKeyLength = 8

// EventLookupKey is a globally unique reference to the event:
// block index + index of the request within block + index of the event within the request
type EventLookupKey [EventLookupKeyLength]byte

func NewEventLookupKey(blockIndex uint32, requestIndex, eventIndex uint16) *EventLookupKey {
	var ret EventLookupKey
	copy(ret[:4], codec.Encode[uint32](blockIndex))
	copy(ret[4:6], codec.Encode[uint16](requestIndex))
	copy(ret[6:8], codec.Encode[uint16](eventIndex))
	return &ret
}

func (k *EventLookupKey) BlockIndex() uint32 {
	return codec.MustDecode[uint32](k[:4])
}

func (k *EventLookupKey) RequestIndex() uint16 {
	return codec.MustDecode[uint16](k[4:6])
}

func (k *EventLookupKey) RequestEventIndex() uint16 {
	return codec.MustDecode[uint16](k[6:8])
}

func (k *EventLookupKey) Bytes() []byte {
	return bcs.MustMarshal(k)
}
