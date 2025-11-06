// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type msgBLSPartialSig struct {
	blsSuite   suites.Suite
	partialSig []byte `bcs:"export"`
}

var _ gpa.MessagePayload = new(msgBLSPartialSig)

func newMsgBLSPartialSig(blsSuite suites.Suite, recipient gpa.NodeID, partialSig []byte) gpa.MessageOut {
	return gpa.NewMessageOut(recipient, &msgBLSPartialSig{
		blsSuite:   blsSuite,
		partialSig: partialSig,
	})
}

func (msg *msgBLSPartialSig) MsgType() gpa.MessageType {
	return msgTypeBLSShare
}
