// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package accessMgr

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

// Send by a node which has a chain enabled to a node it considers an access node.
type msgAccess struct {
	gpa.BasicMessage
	senderLClock   int
	receiverLClock int
	accessToChains []isc.ChainID
}

var _ gpa.Message = &msgAccess{}

func newMsgAccess(
	recipient gpa.NodeID,
	senderLClock, receiverLClock int,
	accessToChains []isc.ChainID,
) gpa.Message {
	return &msgAccess{
		BasicMessage:   gpa.NewBasicMessage(recipient),
		senderLClock:   senderLClock,
		receiverLClock: receiverLClock,
		accessToChains: accessToChains,
	}
}

func (m *msgAccess) MarshalBinary() ([]byte, error) {
	panic("not implemented") // TODO: ..
}

func (m *msgAccess) UnmarshalBinary(data []byte) error {
	panic("not implemented") // TODO: ..
}
