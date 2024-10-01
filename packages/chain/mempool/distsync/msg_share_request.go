// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distsync

import (
	"io"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/bcs"
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

func (msg *msgShareRequest) Equals(other *msgShareRequest) bool {
	return msg.BasicMessage.Equals(&other.BasicMessage) &&
		msg.request.Equals(other.request) &&
		msg.ttl == other.ttl
}

func (msg *msgShareRequest) Read(r io.Reader) error {
	var err error
	rr := rwutil.NewReader(r)
	msgTypeShareRequest.ReadAndVerify(rr)
	msg.ttl = rr.ReadByte()

	msg.request, err = bcs.Unmarshal[isc.Request](rr.ReadBytes())
	if err != nil {
		return err
	}
	return rr.Err
}

func (msg *msgShareRequest) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	msgTypeShareRequest.Write(ww)
	ww.WriteByte(msg.ttl)
	bytes, err := bcs.Marshal(&msg.request)
	if err != nil {
		return nil
	}
	ww.WriteBytes(bytes)
	return ww.Err
}
