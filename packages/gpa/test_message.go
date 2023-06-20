// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"io"

	"github.com/iotaledger/wasp/packages/util/rwutil"
)

const msgTypeTest MessageType = 0xff

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

func (msg *TestMessage) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msgTypeTest.ReadAndVerify(rr)
	msg.ID = int(rr.ReadUint32())
	return rr.Err
}

func (msg *TestMessage) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	msgTypeTest.Write(ww)
	ww.WriteUint32(uint32(msg.ID))
	return ww.Err
}
