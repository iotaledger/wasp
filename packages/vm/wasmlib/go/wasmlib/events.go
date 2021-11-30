// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"strconv"
)

// encodes separate entities into a byte buffer
type EventEncoder struct {
	event string
}

func NewEventEncoder(eventName string) *EventEncoder {
	e := &EventEncoder{event: eventName}
	timestamp := Root.GetInt64(KeyTimestamp).Value()
	// convert nanoseconds to seconds
	return e.Int64(timestamp / 1_000_000_000)
}

func (e *EventEncoder) Address(value ScAddress) *EventEncoder {
	return e.String(value.String())
}

func (e *EventEncoder) AgentID(value ScAgentID) *EventEncoder {
	return e.String(value.String())
}

func (e *EventEncoder) Bool(value bool) *EventEncoder {
	if value {
		return e.Uint8(1)
	}
	return e.Uint8(0)
}

func (e *EventEncoder) Bytes(value []byte) *EventEncoder {
	return e.String(base58Encode(value))
}

func (e *EventEncoder) ChainID(value ScChainID) *EventEncoder {
	return e.String(value.String())
}

func (e *EventEncoder) Color(value ScColor) *EventEncoder {
	return e.String(value.String())
}

func (e *EventEncoder) Emit() {
	Root.GetString(KeyEvent).SetValue(e.event)
}

func (e *EventEncoder) Hash(value ScHash) *EventEncoder {
	return e.String(value.String())
}

func (e *EventEncoder) Hname(value ScHname) *EventEncoder {
	return e.String(value.String())
}

func (e *EventEncoder) Int8(value int8) *EventEncoder {
	return e.Int64(int64(value))
}

func (e *EventEncoder) Int16(value int16) *EventEncoder {
	return e.Int64(int64(value))
}

func (e *EventEncoder) Int32(value int32) *EventEncoder {
	return e.Int64(int64(value))
}

func (e *EventEncoder) Int64(value int64) *EventEncoder {
	return e.String(strconv.FormatInt(value, 10))
}

func (e *EventEncoder) RequestID(value ScRequestID) *EventEncoder {
	return e.String(value.String())
}

func (e *EventEncoder) String(value string) *EventEncoder {
	// TODO encode potential vertical bars that are present in the value string
	e.event += "|" + value
	return e
}

func (e *EventEncoder) Uint8(value uint8) *EventEncoder {
	return e.Uint64(uint64(value))
}

func (e *EventEncoder) Uint16(value uint16) *EventEncoder {
	return e.Uint64(uint64(value))
}

func (e *EventEncoder) Uint32(value uint32) *EventEncoder {
	return e.Uint64(uint64(value))
}

func (e *EventEncoder) Uint64(value uint64) *EventEncoder {
	return e.String(strconv.FormatUint(value, 10))
}
