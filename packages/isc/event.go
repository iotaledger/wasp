package isc

import (
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type Event struct {
	ContractID Hname
	Payload    []byte
	Topic      string
	Timestamp  uint64
}

func NewEvent(data []byte) (*Event, error) {
	return rwutil.ReaderFromBytes(data, new(Event))
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
	return rwutil.WriterToBytes(e)
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

func (e *Event) ToJSONStruct() *EventJSON {
	return &EventJSON{
		ContractID: e.ContractID,
		Payload:    iotago.EncodeHex(e.Payload),
		Topic:      e.Topic,
		Timestamp:  e.Timestamp,
	}
}

type EventJSON struct {
	ContractID Hname  `json:"contractID" swagger:"desc(ID of the Contract that issued the event),required,min(1)"`
	Payload    string `json:"payload" swagger:"desc(payload),required"`
	Topic      string `json:"topic" swagger:"desc(topic),required"`
	Timestamp  uint64 `json:"timestamp" swagger:"desc(timestamp),required"`
}
