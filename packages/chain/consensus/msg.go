// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/chain/distsign"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/gpa/aba/mostefaoui"
	"github.com/iotaledger/wasp/v2/packages/gpa/acs"
	"github.com/iotaledger/wasp/v2/packages/gpa/acss"
	"github.com/iotaledger/wasp/v2/packages/gpa/asyncdistkeygen/nonce"
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
		return gpa.MarshalPayload(msgTypeBLSShare, p)
	case gpa.PayloadWithKey[int, any]:
		switch p.Payload.(type) {
		case mostefaoui.MsgDone:
			return gpa.MarshalPayload(msgTypeABAMsgDone, p)
		case mostefaoui.MsgVote:
			return gpa.MarshalPayload(msgTypeABAMsgVote, p)
		case blssig.MsgSigShare:
			return gpa.MarshalPayload(msgTypeBLSSigShare, p)
		case bracha.MsgBracha:
			switch p.SubsystemID {
			case acs.SubsystemID:
				return gpa.MarshalPayload(msgTypeACSBracha, p)
			case nonce.SubsystemID:
				return gpa.MarshalPayload(msgTypeACSSBracha, p)
			default:
				panic(fmt.Errorf("unexpected subsystem ID: %s", p.SubsystemID))
			}
		case acss.MsgVote:
			return gpa.MarshalPayload(msgTypeACSSVote, p)
		case acss.MsgImplicateRecover:
			return gpa.MarshalPayload(msgTypeACSSImplicateRecover, p)
		default:
			panic(fmt.Errorf("unexpected payload type: %T", p.Payload))
		}
	case distsign.MsgPartialSig:
		return gpa.MarshalPayload(msgTypeDSSPartialSig, p)
	// TODO: Organize this consisntently before merge
	default:
		panic(fmt.Errorf("unexpected payload type: %T", payload))
	}
}

func (c *Consensus) UnmarshalPayload(data []byte) (any, error) {
	return gpa.UnmarshalPayload(data, gpa.PayloadAllocator{
		msgTypeBLSShare:             func() any { return msgBLSPartialSig{} },
		msgTypeACSBracha:            func() any { return gpa.PayloadWithKey[int, bracha.MsgBracha]{} },
		msgTypeABAMsgDone:           func() any { return gpa.PayloadWithKey[int, mostefaoui.MsgDone]{} },
		msgTypeABAMsgVote:           func() any { return gpa.PayloadWithKey[int, mostefaoui.MsgVote]{} },
		msgTypeACSSBracha:           func() any { return gpa.PayloadWithKey[int, bracha.MsgBracha]{} },
		msgTypeACSSVote:             func() any { return gpa.PayloadWithKey[int, acss.MsgVote]{} },
		msgTypeACSSImplicateRecover: func() any { return gpa.PayloadWithKey[int, acss.MsgImplicateRecover]{} },
		msgTypeDSSPartialSig:        func() any { return distsign.MsgPartialSig{} },
		msgTypeBLSSigShare:          func() any { return gpa.PayloadWithKey[int, blssig.MsgSigShare]{} },
	})
}
