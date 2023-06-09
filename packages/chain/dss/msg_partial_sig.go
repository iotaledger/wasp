// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dss

import (
	"io"

	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/dss"
	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type msgPartialSig struct {
	suite      suites.Suite // Transient, for un-marshaling only.
	sender     gpa.NodeID   // Transient.
	recipient  gpa.NodeID   // Transient.
	partialSig *dss.PartialSig
}

var _ gpa.Message = new(msgPartialSig)

func (msg *msgPartialSig) Recipient() gpa.NodeID {
	return msg.recipient
}

func (msg *msgPartialSig) SetSender(sender gpa.NodeID) {
	msg.sender = sender
}

func (msg *msgPartialSig) MarshalBinary() ([]byte, error) {
	return rwutil.WriterToBytes(msg), nil
}

func (msg *msgPartialSig) UnmarshalBinary(data []byte) error {
	_, err := rwutil.ReaderFromBytes(data, msg)
	return err
}

func (msg *msgPartialSig) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.ReadMessageTypeAndVerify(msgTypePartialSig)
	msg.partialSig = &dss.PartialSig{
		Partial: &share.PriShare{
			I: int(rr.ReadUint16()),
			V: msg.suite.Scalar(),
		},
	}
	rr.ReadMarshaled(msg.partialSig.Partial.V)
	msg.partialSig.SessionID = rr.ReadBytes()
	msg.partialSig.Signature = rr.ReadBytes()
	return rr.Err
}

func (msg *msgPartialSig) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteMessageType(msgTypePartialSig)
	ww.WriteUint16(uint16(msg.partialSig.Partial.I)) // TODO: Resolve it from the context, instead of marshaling.
	ww.WriteMarshaled(msg.partialSig.Partial.V)
	ww.WriteBytes(msg.partialSig.SessionID)
	ww.WriteBytes(msg.partialSig.Signature)
	return ww.Err
}
