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

func NewEvent(eventData []byte) (*Event, error) {
	if len(eventData) < 4+2+8 {
		return nil, errors.New("insufficient event data")
	}
	hContract := Hname(binary.LittleEndian.Uint32(eventData[:4]))
	eventData = eventData[4:]
	length := binary.LittleEndian.Uint16(eventData[:2])
	eventData = eventData[2:]
	if len(eventData) < int(length)+8 {
		return nil, errors.New("insufficient event topic data")
	}
	topic := string(eventData[:length])
	eventData = eventData[length:]
	timestamp := binary.LittleEndian.Uint64(eventData[:8])
	return &Event{
		ContractID: hContract,
		Payload:    eventData[8:],
		Timestamp:  timestamp,
		Topic:      topic,
	}, nil
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
