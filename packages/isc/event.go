package isc

import (
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/iotaledger/wasp/packages/util/bcs"
)

type Event struct {
	ContractID Hname  `json:"contractID"`
	Topic      string `json:"topic"`
	Timestamp  uint64 `json:"timestamp"`
	Payload    []byte `json:"payload"`
}

func EventFromBytes(data []byte) (*Event, error) {
	return bcs.Unmarshal[*Event](data)
}

// ContractIDFromEventBytes is used by blocklog to filter out specific events per contract
// For performance reasons it is working directly with the event bytes.
func ContractIDFromEventBytes(eventBytes []byte) (Hname, error) {
	if len(eventBytes) > 4 {
		return HnameFromBytes(eventBytes[:4])
	}

	return HnameNil, nil
}

func (e *Event) Bytes() []byte {
	return bcs.MustMarshal(e)
}

type EventJSON struct {
	ContractID Hname  `json:"contractID" swagger:"desc(ID of the Contract that issued the event),required,min(1)"`
	Topic      string `json:"topic" swagger:"desc(topic),required"`
	Timestamp  uint64 `json:"timestamp" swagger:"desc(timestamp),required"`
	Payload    string `json:"payload" swagger:"desc(payload),required"`
}

func (e *Event) ToJSONStruct() *EventJSON {
	return &EventJSON{
		ContractID: e.ContractID,
		Topic:      e.Topic,
		Timestamp:  e.Timestamp,
		Payload:    hexutil.Encode(e.Payload),
	}
}
