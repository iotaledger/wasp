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
	msgTypeACSBracha
	msgTypeABAMsgDone
	msgTypeABAMsgVote
	msgTypeACSSBracha
	msgTypeACSSVote
	msgTypeACSSImplicateRecover
	msgTypeDSSPartialSig
	msgTypeBLSSigShare
)

func (c *Consensus) MarshalPayload(payload any) ([]byte, error) {
	switch p := payload.(type) {
	case msgBLSPartialSig:
		return gpa.MarshalPayloadNEW(msgTypeBLSShare, p)
	case gpa.PayloadWithKey[int, any]:
		switch p.Payload.(type) {
		case bracha.MsgBracha:
			return gpa.MarshalPayloadNEW(msgTypeACSBracha, p)
		case mostefaoui.MsgDone:
			return gpa.MarshalPayloadNEW(msgTypeABAMsgDone, p)
		case mostefaoui.MsgVote:
			return gpa.MarshalPayloadNEW(msgTypeABAMsgVote, p)
		case blssig.MsgSigShare:
			return gpa.MarshalPayloadNEW(msgTypeBLSSigShare, p)
		default:
			panic(fmt.Errorf("unexpected payload type: %T", p.Payload))
		}
	case distsign.MsgPartialSig:
		return gpa.MarshalPayloadNEW(msgTypeDSSPartialSig, p)
	// TODO: Organize this consisntently before merge
	case gpa.SubsystemPayload[int, any]:
		switch p.Payload.(type) {
		case bracha.MsgBracha:
			switch p.SubsystemID {
			case "acss":
				return gpa.MarshalPayloadNEW(msgTypeACSSBracha, p)
			default:
				panic(fmt.Errorf("unexpected subsystem ID: %s", p.SubsystemID))
			}
		case acss.MsgVote:
			return gpa.MarshalPayloadNEW(msgTypeACSSVote, p)
		case acss.MsgImplicateRecover:
			return gpa.MarshalPayloadNEW(msgTypeACSSImplicateRecover, p)
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
		msgTypeACSBracha:            func() any { return gpa.PayloadWithKey[int, bracha.MsgBracha]{} },
		msgTypeABAMsgDone:           func() any { return gpa.PayloadWithKey[int, mostefaoui.MsgDone]{} },
		msgTypeABAMsgVote:           func() any { return gpa.PayloadWithKey[int, mostefaoui.MsgVote]{} },
		msgTypeACSSBracha:           func() any { return gpa.SubsystemPayload[int, bracha.MsgBracha]{} },
		msgTypeACSSVote:             func() any { return gpa.SubsystemPayload[int, acss.MsgVote]{} },
		msgTypeACSSImplicateRecover: func() any { return gpa.SubsystemPayload[int, acss.MsgImplicateRecover]{} },
		msgTypeDSSPartialSig:        func() any { return distsign.MsgPartialSig{} },
		msgTypeBLSSigShare:          func() any { return gpa.PayloadWithKey[int, blssig.MsgSigShare]{} },
	})
}
