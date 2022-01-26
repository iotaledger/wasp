// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"strconv"

	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmtypes"
)

// encodes separate entities into a byte buffer
type EventEncoder struct {
	event string
}

func NewEventEncoder(eventName string) *EventEncoder {
	e := &EventEncoder{event: eventName}
	timestamp := ScSandbox{}.Timestamp()
	// convert nanoseconds to seconds
	e.Encode(wasmtypes.Uint64ToString(timestamp / 1_000_000_000))
	return e
}

func (e *EventEncoder) Address(value wasmtypes.ScAddress) *EventEncoder {
	return e.String(value.String())
}

func (e *EventEncoder) AgentID(value wasmtypes.ScAgentID) *EventEncoder {
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

func (e *EventEncoder) ChainID(value wasmtypes.ScChainID) *EventEncoder {
	return e.String(value.String())
}

func (e *EventEncoder) Color(value wasmtypes.ScColor) *EventEncoder {
	return e.String(value.String())
}

func (e *EventEncoder) Emit() {
	ScSandboxFunc{}.Event(e.event)
}

func (e *EventEncoder) Encode(value string) {
	// TODO encode potential vertical bars that are present in the value string
	e.event += "|" + value
}

func (e *EventEncoder) Hash(value wasmtypes.ScHash) *EventEncoder {
	return e.String(value.String())
}

func (e *EventEncoder) Hname(value wasmtypes.ScHname) *EventEncoder {
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

func (e *EventEncoder) RequestID(value wasmtypes.ScRequestID) *EventEncoder {
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
