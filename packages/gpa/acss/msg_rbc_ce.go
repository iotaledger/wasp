// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"io"

	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

// This message is used as a payload of the RBC:
//
// > RBC(C||E)
type msgRBCCEPayload struct {
	gpa.BasicMessage
	suite suites.Suite
	data  []byte
	err   error // Transient field, should not be serialized.
}

var _ gpa.Message = new(msgRBCCEPayload)

func (msg *msgRBCCEPayload) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msgTypeRBCCEPayload.ReadAndVerify(rr)
	msg.data = rr.ReadBytes()
	return rr.Err
}

func (msg *msgRBCCEPayload) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	msgTypeRBCCEPayload.Write(ww)
	ww.WriteBytes(msg.data)
	return ww.Err
}
