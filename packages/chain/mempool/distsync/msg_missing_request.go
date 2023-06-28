// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distsync

import (
	"io"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type msgMissingRequest struct {
	gpa.BasicMessage
	requestRef *isc.RequestRef
}

var _ gpa.Message = new(msgMissingRequest)

func newMsgMissingRequest(requestRef *isc.RequestRef, recipient gpa.NodeID) gpa.Message {
	return &msgMissingRequest{
		BasicMessage: gpa.NewBasicMessage(recipient),
		requestRef:   requestRef,
	}
}

func (msg *msgMissingRequest) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msgTypeMissingRequest.ReadAndVerify(rr)
	msg.requestRef = new(isc.RequestRef)
	rr.Read(msg.requestRef)
	return rr.Err
}

func (msg *msgMissingRequest) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	msgTypeMissingRequest.Write(ww)
	ww.Write(msg.requestRef)
	return ww.Err
}
