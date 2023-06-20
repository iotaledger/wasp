// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"io"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type msgVoteType byte

const (
	BVAL msgVoteType = iota
	AUX
)

type msgVote struct {
	gpa.BasicMessage
	round    int
	voteType msgVoteType
	value    bool
}

var _ gpa.Message = new(msgVote)

func multicastMsgVote(recipients []gpa.NodeID, round int, voteType msgVoteType, value bool) gpa.OutMessages {
	msgs := gpa.NoMessages()
	for _, recipient := range recipients {
		msgs.Add(&msgVote{
			BasicMessage: gpa.NewBasicMessage(recipient),
			round:        round,
			voteType:     voteType,
			value:        value,
		})
	}
	return msgs
}

func (msg *msgVote) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msgTypeVote.ReadAndVerify(rr)
	msg.round = int(rr.ReadUint16())
	msg.voteType = msgVoteType(rr.ReadByte())
	msg.value = rr.ReadBool()
	return rr.Err
}

func (msg *msgVote) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	msgTypeVote.Write(ww)
	ww.WriteUint16(uint16(msg.round))
	ww.WriteByte(byte(msg.voteType))
	ww.WriteBool(msg.value)
	return ww.Err
}
