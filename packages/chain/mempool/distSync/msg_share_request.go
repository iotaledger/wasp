// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distSync

import (
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type msgShareRequest struct {
	gpa.BasicMessage
	request isc.Request
	ttl     byte
}

var _ gpa.Message = &msgShareRequest{}

func newMsgShareRequest(request isc.Request, ttl byte, recipient gpa.NodeID) gpa.Message {
	return &msgShareRequest{
		BasicMessage: gpa.NewBasicMessage(recipient),
		request:      request,
		ttl:          ttl,
	}
}

func (msg *msgShareRequest) MarshalBinary() (data []byte, err error) {
	ret := []byte{msgTypeShareRequest, msg.ttl}
	ret = append(ret, msg.request.Bytes()...)
	return ret, nil
}

func (msg *msgShareRequest) UnmarshalBinary(data []byte) (err error) {
	if len(data) < 2 {
		return xerrors.Errorf("cannot parse a message, data to short, len=%v", len(data))
	}
	if data[0] != msgTypeShareRequest {
		return xerrors.Errorf("cannot parse a message, unexpected msgType=%v", data[0])
	}
	msg.ttl = data[1]
	msg.request, err = isc.NewRequestFromBytes(data[2:])
	return err
}
