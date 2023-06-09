// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package blssig

import (
	"io"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type msgSigShare struct {
	gpa.BasicMessage
	sigShare []byte
}

var _ gpa.Message = new(msgSigShare)

func (msg *msgSigShare) MarshalBinary() ([]byte, error) {
	return rwutil.MarshalBinary(msg)
}

func (msg *msgSigShare) UnmarshalBinary(data []byte) error {
	return rwutil.UnmarshalBinary(data, msg)
}

func (msg *msgSigShare) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.ReadMessageTypeAndVerify(msgTypeSigShare)
	msg.sigShare = rr.ReadBytes()
	return rr.Err
}

func (msg *msgSigShare) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteMessageType(msgTypeSigShare)
	ww.WriteBytes(msg.sigShare)
	return ww.Err
}
