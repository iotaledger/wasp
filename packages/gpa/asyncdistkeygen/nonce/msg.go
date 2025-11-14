// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nonce

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/gpa/acss"
	rbc "github.com/iotaledger/wasp/v2/packages/gpa/rbc/bracha"
)

type MsgACSSVote struct {
	Index int
	acss.MsgVote
}

type MsgACSSRBCCEPayload struct {
	Index int
	acss.MsgRBCCEPayload
}

type MsgACSSImplicateRecover struct {
	Index int
	acss.MsgImplicateRecover
}

type MsgACSSBracha struct {
	Index int
	rbc.MsgBracha
}

type OutMessages *OutMessagesV
type OutMessagesV struct {
	Vote             []gpa.TypedPayloadOut[MsgACSSVote]
	RBCCEPayload     []gpa.TypedPayloadOut[MsgACSSRBCCEPayload]
	ImplicateRecover []gpa.TypedPayloadOut[MsgACSSImplicateRecover]
	Bracha           []gpa.TypedPayloadOut[MsgACSSBracha]
}

func (m *OutMessagesV) AddAll(msgs OutMessages) OutMessages {
	if msgs != nil {
		m.Vote = append(m.Vote, msgs.Vote...)
		m.RBCCEPayload = append(m.RBCCEPayload, msgs.RBCCEPayload...)
		m.ImplicateRecover = append(m.ImplicateRecover, msgs.ImplicateRecover...)
		m.Bracha = append(m.Bracha, msgs.Bracha...)
	}
	return m
}

func NoMessages() *OutMessagesV {
	return &OutMessagesV{}
}

func ConcatMsgs(msgs ...OutMessages) OutMessages {
	res := NoMessages()
	for _, m := range msgs {
		res.AddAll(m)
	}
	return res
}
