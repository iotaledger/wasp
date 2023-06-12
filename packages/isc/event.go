package isc

import (
	"io"

	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type Event struct {
	ContractID Hname  `json:"contractID"`
	Payload    []byte `json:"payload"`
	Topic      string `json:"topic"`
	Timestamp  uint64 `json:"timestamp"`
}

func NewEvent(data []byte) (*Event, error) {
	return rwutil.ReaderFromBytes(data, new(Event))
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
