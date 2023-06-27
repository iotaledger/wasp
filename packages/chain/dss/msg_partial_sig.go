// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dss

import (
	"io"

	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/dss"
	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type msgPartialSig struct {
	gpa.BasicMessage
	suite      suites.Suite // Transient, for un-marshaling only.
	partialSig *dss.PartialSig
}

var _ gpa.Message = new(msgPartialSig)

func (msg *msgPartialSig) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msgTypePartialSig.ReadAndVerify(rr)
	msg.partialSig = &dss.PartialSig{Partial: &share.PriShare{}}
	msg.partialSig.Partial.I = int(rr.ReadUint16())
	msg.partialSig.Partial.V = cryptolib.ScalarFromReader(rr, msg.suite)
	msg.partialSig.SessionID = rr.ReadBytes()
	msg.partialSig.Signature = rr.ReadBytes()
	return rr.Err
}

func (msg *msgPartialSig) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	msgTypePartialSig.Write(ww)
	ww.WriteUint16(uint16(msg.partialSig.Partial.I)) // TODO: Resolve it from the context, instead of marshaling.
	cryptolib.ScalarToWriter(ww, msg.partialSig.Partial.V)
	ww.WriteBytes(msg.partialSig.SessionID)
	ww.WriteBytes(msg.partialSig.Signature)
	return ww.Err
}
