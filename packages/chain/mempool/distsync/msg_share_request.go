// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distsync

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type msgShareRequest struct {
	gpa.BasicMessage
	ttl     byte        `bcs:"export"`
	request isc.Request `bcs:"export"`
}

var _ gpa.Message = new(msgShareRequest)

func newMsgShareRequest(request isc.Request, ttl byte, recipient gpa.NodeID) gpa.Message {
	return &msgShareRequest{
		BasicMessage: gpa.NewBasicMessage(recipient),
		request:      request,
		ttl:          ttl,
	}
}

func (msg *msgShareRequest) MsgType() gpa.MessageType {
	return msgTypeShareRequest
}
