package isc

import (
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type Event struct {
	ContractID Hname  `json:"contractID"`
	Topic      string `json:"topic"`
	Timestamp  uint64 `json:"timestamp"`
	Payload    []byte `json:"payload"`
}

func EventFromBytes(data []byte) (*Event, error) {
	return rwutil.ReadFromBytes(data, new(Event))
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
	return rwutil.WriteToBytes(e)
}

func (e *Event) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.Read(&e.ContractID)
	e.Topic = rr.ReadString()
	e.Timestamp = rr.ReadUint64()
	e.Payload = rr.ReadBytes()
	return rr.Err
}

func (e *Event) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.Write(&e.ContractID)
	ww.WriteString(e.Topic)
	ww.WriteUint64(e.Timestamp)
	ww.WriteBytes(e.Payload)
	return ww.Err
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
		Payload:    iotago.EncodeHex(e.Payload),
	}
}
