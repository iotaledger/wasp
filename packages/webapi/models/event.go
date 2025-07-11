package models

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iotaledger/wasp/packages/isc"
)

type EventJSON struct {
	ContractID isc.Hname `json:"contractID" swagger:"desc(ID of the Contract that issued the event),required,min(1)"`
	Topic      string    `json:"topic" swagger:"desc(topic),required"`
	Timestamp  uint64    `json:"timestamp" swagger:"desc(timestamp),required"`
	Payload    string    `json:"payload" swagger:"desc(payload),required"`
}

func ToJSONStruct(e *isc.Event) *EventJSON {
	return &EventJSON{
		ContractID: e.ContractID,
		Topic:      e.Topic,
		Timestamp:  e.Timestamp,
		Payload:    hexutil.Encode(e.Payload),
	}
}
