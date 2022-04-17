// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type EventEncoder struct {
	event string
}

func NewEventEncoder(eventName string) *EventEncoder {
	e := &EventEncoder{event: eventName}
	timestamp := ScFuncContext{}.Timestamp()
	// convert nanoseconds to seconds
	e.Encode(wasmtypes.Uint64ToString(timestamp / 1_000_000_000))
	return e
}

func (e *EventEncoder) Emit() {
	ScFuncContext{}.Event(e.event)
}

func (e *EventEncoder) Encode(value string) {
	// TODO encode potential vertical bars that are present in the value string
	e.event += "|" + value
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type EventDecoder struct {
	msg []string
}

func NewEventDecoder(msg []string) *EventDecoder {
	return &EventDecoder{msg: msg}
}

func (d *EventDecoder) Decode() string {
	next := d.msg[0]
	d.msg = d.msg[1:]
	return next
}

func (d *EventDecoder) Timestamp() uint32 {
	return wasmtypes.Uint32FromString(d.Decode())
}
