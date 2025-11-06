// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distsync

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type msgMissingRequest struct {
	requestRef *isc.RequestRef `bcs:"export"`
}

var _ gpa.MessagePayload = new(msgMissingRequest)

func newMsgMissingRequest(requestRef *isc.RequestRef, recipient gpa.NodeID) gpa.MessageOut {
	return gpa.NewMessageOut(recipient, &msgMissingRequest{
		requestRef: requestRef,
	})
}

func (msg *msgMissingRequest) MsgType() gpa.MessageType {
	return msgTypeMissingRequest
}
