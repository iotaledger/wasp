// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/chain/distsign"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/gpa/aba/mostefaoui"
	"github.com/iotaledger/wasp/v2/packages/gpa/acss"
	"github.com/iotaledger/wasp/v2/packages/gpa/cc/blssig"
	"github.com/iotaledger/wasp/v2/packages/gpa/rbc/bracha"
)

const (
	msgTypeBLSShare gpa.MessageType = iota
	msgTypeRBCBracha
	msgTypeABAMsgDone
	msgTypeABAMsgVote
	msgTypeACSSVote
	msgTypeACSSImplicateRecover
	msgTypeDSSPartialSig
	msgTypeBLSSigShare
)

func (c *Consensus) MarshalPayload(payload any) ([]byte, error) {
	switch p := payload.(type) {
	case msgBLSPartialSig:
		return gpa.MarshalPayloadNEW(msgTypeBLSShare, p)
	case gpa.PayloadWithIndex[any]:
		switch p.Payload.(type) {
		case bracha.MsgBracha:
			return gpa.MarshalPayloadNEW(msgTypeRBCBracha, p)
		case mostefaoui.MsgDone:
			return gpa.MarshalPayloadNEW(msgTypeABAMsgDone, p)
		case mostefaoui.MsgVote:
			return gpa.MarshalPayloadNEW(msgTypeABAMsgVote, p)
		case acss.MsgVote:
			return gpa.MarshalPayloadNEW(msgTypeACSSVote, p)
		case acss.MsgImplicateRecover:
			return gpa.MarshalPayloadNEW(msgTypeACSSImplicateRecover, p)
		case distsign.MsgPartialSig:
			return gpa.MarshalPayloadNEW(msgTypeDSSPartialSig, p)
		case blssig.MsgSigShare:
			return gpa.MarshalPayloadNEW(msgTypeBLSSigShare, p)
		default:
			panic(fmt.Errorf("unexpected payload type: %T", p.Payload))
		}
	default:
		panic(fmt.Errorf("unexpected payload type: %T", payload))
	}
}

func (c *Consensus) UnmarshalPayload(data []byte) (any, error) {
	return gpa.UnmarshalPayloadNEW(data, gpa.PayloadAllocatorNEW{
		msgTypeBLSShare:             func() any { return msgBLSPartialSig{} },
		msgTypeRBCBracha:            func() any { return gpa.PayloadWithIndex[bracha.MsgBracha]{} },
		msgTypeABAMsgDone:           func() any { return gpa.PayloadWithIndex[mostefaoui.MsgDone]{} },
		msgTypeABAMsgVote:           func() any { return gpa.PayloadWithIndex[mostefaoui.MsgVote]{} },
		msgTypeACSSVote:             func() any { return gpa.PayloadWithIndex[acss.MsgVote]{} },
		msgTypeACSSImplicateRecover: func() any { return gpa.PayloadWithIndex[acss.MsgImplicateRecover]{} },
		msgTypeDSSPartialSig:        func() any { return gpa.PayloadWithIndex[distsign.MsgPartialSig]{} },
		msgTypeBLSSigShare:          func() any { return gpa.PayloadWithIndex[blssig.MsgSigShare]{} },
	})
}
