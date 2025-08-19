// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dist

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

// Send by a node which has a chain enabled to a node it considers an access node.
type msgAccess struct {
	gpa.BasicMessage
	senderLClock   int  `bcs:"export,type=u32"`
	receiverLClock int  `bcs:"export,type=u32"`
	isAccessNode   bool `bcs:"export"`
	isServer       bool `bcs:"export"`
}

var _ gpa.Message = new(msgAccess)

func newMsgAccess(
	recipient gpa.NodeID,
	senderLClock, receiverLClock int,
	isAccessNode bool,
	isServer bool,
) gpa.Message {
	return &msgAccess{
		BasicMessage:   gpa.NewBasicMessage(recipient),
		senderLClock:   senderLClock,
		receiverLClock: receiverLClock,
		isAccessNode:   isAccessNode,
		isServer:       isServer,
	}
}

func (msg *msgAccess) MsgType() gpa.MessageType {
	return msgTypeAccess
}
