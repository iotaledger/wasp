// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type msgBLSPartialSig struct {
	gpa.BasicMessage
	blsSuite   suites.Suite
	partialSig []byte `bcs:"export"`
}

var _ gpa.Message = new(msgBLSPartialSig)

func newMsgBLSPartialSig(blsSuite suites.Suite, recipient gpa.NodeID, partialSig []byte) *msgBLSPartialSig {
	return &msgBLSPartialSig{
		BasicMessage: gpa.NewBasicMessage(recipient),
		blsSuite:     blsSuite,
		partialSig:   partialSig,
	}
}

func (msg *msgBLSPartialSig) MsgType() gpa.MessageType {
	return msgTypeBLSShare
}
