package isc

import (
	bcs "github.com/iotaledger/bcs-go"
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
