// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"io"

	"github.com/iotaledger/wasp/packages/util/rwutil"
	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/wasp/packages/gpa"
)

type msgBLSPartialSig struct {
	blsSuite   suites.Suite
	sender     gpa.NodeID
	recipient  gpa.NodeID
	partialSig []byte
}

var _ gpa.Message = new(msgBLSPartialSig)

func newMsgBLSPartialSig(blsSuite suites.Suite, recipient gpa.NodeID, partialSig []byte) *msgBLSPartialSig {
	return &msgBLSPartialSig{blsSuite: blsSuite, recipient: recipient, partialSig: partialSig}
}

func (msg *msgBLSPartialSig) Recipient() gpa.NodeID {
	return msg.recipient
}

func (msg *msgBLSPartialSig) SetSender(sender gpa.NodeID) {
	msg.sender = sender
}

func (msg *msgBLSPartialSig) MarshalBinary() ([]byte, error) {
	return rwutil.WriterToBytes(msg), nil
}

func (msg *msgBLSPartialSig) UnmarshalBinary(data []byte) error {
	_, err := rwutil.ReaderFromBytes(data, msg)
	return err
}

func (msg *msgBLSPartialSig) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.ReadMessageTypeAndVerify(msgTypeBLSShare)
	msg.partialSig = rr.ReadBytes()
	return rr.Err
}

func (msg *msgBLSPartialSig) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteMessageType(msgTypeBLSShare)
	ww.WriteBytes(msg.partialSig)
	return ww.Err
}
