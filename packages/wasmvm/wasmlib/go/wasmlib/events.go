// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"strings"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type EventEncoder struct {
	event string
}

func NewEventEncoder(eventName string) *EventEncoder {
	e := &EventEncoder{event: eventName}
	e.Encode(wasmtypes.Uint64ToString(ScFuncContext{}.Timestamp()))
	return e
}

func (e *EventEncoder) Emit() {
	ScFuncContext{}.Event(e.event)
}

func (e *EventEncoder) Encode(value string) {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "|", "\\/")
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

func (d *EventDecoder) Timestamp() uint64 {
	return wasmtypes.Uint64FromString(d.Decode())
}
