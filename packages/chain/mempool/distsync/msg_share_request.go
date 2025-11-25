// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distsync

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type msgShareRequest struct {
	ttl     byte        `bcs:"export"`
	request isc.Request `bcs:"export"`
}

func newMsgShareRequest(request isc.Request, ttl byte, recipient gpa.NodeID) gpa.MessageOut {
	return gpa.NewMessageOut(recipient, msgShareRequest{
		request: request,
		ttl:     ttl,
	})
}
