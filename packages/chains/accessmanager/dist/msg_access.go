// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dist

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

// Send by a node which has a chain enabled to a node it considers an access node.
type msgAccess struct {
	senderLClock    int           `bcs:"export,type=u32"`
	receiverLClock  int           `bcs:"export,type=u32"`
	accessForChains []isc.ChainID `bcs:"export,len_bytes=2"`
	serverForChains []isc.ChainID `bcs:"export,len_bytes=2"`
}

var _ gpa.MessagePayload = new(msgAccess)

func newMsgAccess(
	recipient gpa.NodeID,
	senderLClock, receiverLClock int,
	accessForChains []isc.ChainID,
	serverForChains []isc.ChainID,
) *gpa.MessageOut {
	return gpa.NewMessageOut(recipient, &msgAccess{
		senderLClock:    senderLClock,
		receiverLClock:  receiverLClock,
		accessForChains: accessForChains,
		serverForChains: serverForChains,
	})
}

func (msg *msgAccess) MsgType() gpa.MessageType {
	return msgTypeAccess
}
