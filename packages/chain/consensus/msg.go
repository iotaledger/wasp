// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"github.com/iotaledger/wasp/v2/packages/chain/distsign"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/gpa/acss"
	"github.com/iotaledger/wasp/v2/packages/gpa/rbc/bracha"
)

const (
	msgTypeBLSShare gpa.MessageType = iota
	msgTypeRBCBracha
	msgTypeACSSVote
	msgTypeACSSImplicateRecover
	msgTypeDSSPartialSig
)

func (c *Consensus) MarshalPayload(payload any) ([]byte, error) {
	switch p := payload.(type) {
	case msgBLSPartialSig:
		return gpa.MarshalPayloadNEW(msgTypeBLSShare, p)
	case gpa.PayloadWithIndex[bracha.MsgBracha]:
		return gpa.MarshalPayloadNEW(msgTypeRBCBracha, p)
	case gpa.PayloadWithIndex[acss.MsgVote]:
		return gpa.MarshalPayloadNEW(msgTypeACSSVote, p)
	case gpa.PayloadWithIndex[acss.MsgImplicateRecover]:
		return gpa.MarshalPayloadNEW(msgTypeACSSImplicateRecover, p)
	case gpa.PayloadWithIndex[distsign.MsgPartialSig]:
		return gpa.MarshalPayloadNEW(msgTypeDSSPartialSig, p)
	default:
		panic("unexpected payload type")
	}
}

func (c *Consensus) UnmarshalPayload(data []byte) (any, error) {
	return gpa.UnmarshalPayloadNEW(data, gpa.PayloadAllocatorNEW{
		msgTypeBLSShare:             func() any { return &msgBLSPartialSig{blsSuite: c.blsSuite} },
		msgTypeRBCBracha:            func() any { return &gpa.PayloadWithIndex[bracha.MsgBracha]{} },
		msgTypeACSSVote:             func() any { return &gpa.PayloadWithIndex[acss.MsgVote]{} },
		msgTypeACSSImplicateRecover: func() any { return &gpa.PayloadWithIndex[acss.MsgImplicateRecover]{} },
		msgTypeDSSPartialSig:        func() any { return &gpa.PayloadWithIndex[distsign.MsgPartialSig]{} },
	})
}
