// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/chain/committeelog"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

const (
	msgTypeMsgNextLogIndex gpa.MessageType = iota
	msgTypeBlockProduced
)

func (cmi *ChainMgr) MarshalPayload(payload any) ([]byte, error) {
	switch p := payload.(type) {
	case gpa.PayloadWithKey[cryptolib.Address, any]:
		switch p.Payload.(type) {
		case committeelog.MsgNextLogIndex:
			return gpa.MarshalPayloadNEW(msgTypeMsgNextLogIndex, p)
		default:
			panic(fmt.Errorf("chainMgr: unexpected payload type: %T", p.Payload))
		}
	case msgBlockProduced:
		return gpa.MarshalPayloadNEW(msgTypeBlockProduced, p)
	default:
		panic(fmt.Errorf("chainMgr: unknown payload type %T", payload))
	}
}

func (cmi *ChainMgr) UnmarshalPayload(data []byte) (any, error) {
	return gpa.UnmarshalPayloadNEW(data, gpa.PayloadAllocatorNEW{
		msgTypeMsgNextLogIndex: func() any { return gpa.PayloadWithKey[cryptolib.Address, committeelog.MsgNextLogIndex]{} },
		msgTypeBlockProduced:   func() any { return msgBlockProduced{} },
	})
}
