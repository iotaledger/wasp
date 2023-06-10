// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"io"

	"github.com/iotaledger/wasp/packages/util/rwutil"
)

const msgTypeTest = 0xff

// Just a message for test cases.
type TestMessage struct {
	recipient NodeID
	sender    NodeID
	ID        int
}

var _ Message = new(TestMessage)

func (msg *TestMessage) Recipient() NodeID {
	return msg.recipient
}

func (msg *TestMessage) SetSender(sender NodeID) {
	msg.sender = sender
}

func (msg *TestMessage) MarshalBinary() ([]byte, error) {
	return rwutil.MarshalBinary(msg)
}

func (msg *TestMessage) UnmarshalBinary(data []byte) error {
	return rwutil.UnmarshalBinary(data, msg)
}

func (msg *TestMessage) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.ReadKindAndVerify(msgTypeTest)
	msg.ID = int(rr.ReadUint32())
	return rr.Err
}

func (msg *TestMessage) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteKind(msgTypeTest)
	ww.WriteUint32(uint32(msg.ID))
	return ww.Err
}
