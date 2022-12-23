// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package amDist

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

// Send by a node which has a chain enabled to a node it considers an access node.
type msgAccess struct {
	gpa.BasicMessage
	senderLClock    int
	receiverLClock  int
	accessForChains []isc.ChainID
	serverForChains []isc.ChainID
}

var _ gpa.Message = &msgAccess{}

func newMsgAccess(
	recipient gpa.NodeID,
	senderLClock, receiverLClock int,
	accessForChains []isc.ChainID,
	serverForChains []isc.ChainID,
) gpa.Message {
	return &msgAccess{
		BasicMessage:    gpa.NewBasicMessage(recipient),
		senderLClock:    senderLClock,
		receiverLClock:  receiverLClock,
		accessForChains: accessForChains,
		serverForChains: serverForChains,
	}
}

func (m *msgAccess) MarshalBinary() ([]byte, error) {
	panic("not implemented") // TODO: ..
}

func (m *msgAccess) UnmarshalBinary(data []byte) error {
	panic("not implemented") // TODO: ..
}
