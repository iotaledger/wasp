package isc

import (
	"encoding/binary"
	"errors"

	"github.com/iotaledger/wasp/packages/util"
)

type Event struct {
	ContractID Hname  `json:"contractID"`
	Payload    []byte `json:"payload"`
	Topic      string `json:"topic"`
	Timestamp  uint64 `json:"timestamp"`
}

func NewEvent(event []byte) (*Event, error) {
	if len(event) < 4+2+8 {
		return nil, errors.New("insufficient event data")
	}
	hContract := Hname(binary.LittleEndian.Uint32(event[:4]))
	event = event[4:]
	length := binary.LittleEndian.Uint16(event[:2])
	event = event[2:]
	if len(event) < int(length)+8 {
		return nil, errors.New("insufficient event topic data")
	}
	topic := string(event[:length])
	event = event[length:]
	timestamp := binary.LittleEndian.Uint64(event[:8])
	return &Event{
		ContractID: hContract,
		Payload:    event[8:],
		Timestamp:  timestamp,
		Topic:      topic,
	}, nil
}

func NewEvents(events [][]byte) ([]*Event, error) {
	ret := make([]*Event, 0, len(events))
	for _, e := range events {
		event, err := NewEvent(e)
		if err != nil {
			return nil, err
		}
		ret = append(ret, event)
	}
	return ret, nil
}

func (e *Event) Bytes() []byte {
	eventData := make([]byte, 0, 4+2+len(e.Topic)+8+len(e.Payload))
	eventData = append(eventData, e.ContractID.Bytes()...)
	eventData = append(eventData, util.Uint16To2Bytes(uint16(len(e.Topic)))...)
	eventData = append(eventData, []byte(e.Topic)...)
	eventData = append(eventData, util.Uint64To8Bytes(e.Timestamp)...)
	eventData = append(eventData, e.Payload...)
	return eventData
}
