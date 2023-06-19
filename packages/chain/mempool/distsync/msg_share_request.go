// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distsync

import (
	"io"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type msgShareRequest struct {
	gpa.BasicMessage
	request isc.Request
	ttl     byte
}

var _ gpa.Message = new(msgShareRequest)

func newMsgShareRequest(request isc.Request, ttl byte, recipient gpa.NodeID) gpa.Message {
	return &msgShareRequest{
		BasicMessage: gpa.NewBasicMessage(recipient),
		request:      request,
		ttl:          ttl,
	}
}

func (msg *msgShareRequest) MarshalBinary() (data []byte, err error) {
	return rwutil.MarshalBinary(msg)
}

func (msg *msgShareRequest) UnmarshalBinary(data []byte) (err error) {
	return rwutil.UnmarshalBinary(data, msg)
}

func (msg *msgShareRequest) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msgTypeShareRequest.ReadAndVerify(rr)
	msg.ttl = rr.ReadByte()
	msg.request = rwutil.ReadFromBytes(rr, isc.RequestFromBytes)
	return rr.Err
}

func (msg *msgShareRequest) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	msgTypeShareRequest.Write(ww)
	ww.WriteByte(msg.ttl)
	ww.WriteFromBytes(msg.request)
	return ww.Err
}
