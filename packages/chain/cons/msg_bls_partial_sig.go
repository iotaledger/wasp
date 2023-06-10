// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"io"

	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type msgBLSPartialSig struct {
	gpa.BasicMessage
	blsSuite   suites.Suite
	partialSig []byte
}

var _ gpa.Message = new(msgBLSPartialSig)

func newMsgBLSPartialSig(blsSuite suites.Suite, recipient gpa.NodeID, partialSig []byte) *msgBLSPartialSig {
	return &msgBLSPartialSig{
		BasicMessage: gpa.NewBasicMessage(recipient),
		blsSuite:     blsSuite,
		partialSig:   partialSig,
	}
}

func (msg *msgBLSPartialSig) MarshalBinary() ([]byte, error) {
	return rwutil.MarshalBinary(msg)
}

func (msg *msgBLSPartialSig) UnmarshalBinary(data []byte) error {
	return rwutil.UnmarshalBinary(data, msg)
}

func (msg *msgBLSPartialSig) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msgTypeBLSShare.ReadAndVerify(rr)
	msg.partialSig = rr.ReadBytes()
	return rr.Err
}

func (msg *msgBLSPartialSig) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	msgTypeBLSShare.Write(ww)
	ww.WriteBytes(msg.partialSig)
	return ww.Err
}
