// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log

import (
	"fmt"
	"io"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type MsgNextLogIndexCause byte

func (c MsgNextLogIndexCause) String() string {
	switch c {
	case MsgNextLogIndexCauseConsOut:
		return "ConsOut"
	case MsgNextLogIndexCauseRecover:
		return "Recover"
	case MsgNextLogIndexCauseL1RepAO:
		return "L1RepAO"
	case MsgNextLogIndexCauseStarted:
		return "Started"
	default:
		return fmt.Sprintf("%v", byte(c))
	}
}

const (
	MsgNextLogIndexCauseConsOut MsgNextLogIndexCause = iota // Consensus output received, we can go to the next log index.
	MsgNextLogIndexCauseL1RepAO                             // L1 replaced an alias output, probably have to start new log index.
	MsgNextLogIndexCauseRecover                             // Either node is booted or consensus asks for a recovery, try to proceed to next li.
	MsgNextLogIndexCauseStarted                             // Consensus is started, maybe we have to catch up with it.
)

type MsgNextLogIndex struct {
	gpa.BasicMessage
	NextLogIndex LogIndex             // Proposal is to go to this LI without waiting for a consensus.
	Cause        MsgNextLogIndexCause // Reason for the proposal.
	PleaseRepeat bool                 // If true, the receiver should resend its latest message back to the sender.
}

var _ gpa.Message = new(MsgNextLogIndex)

func NewMsgNextLogIndex(recipient gpa.NodeID, nextLogIndex LogIndex, cause MsgNextLogIndexCause, pleaseRepeat bool) *MsgNextLogIndex {
	return &MsgNextLogIndex{
		BasicMessage: gpa.NewBasicMessage(recipient),
		NextLogIndex: nextLogIndex,
		Cause:        cause,
		PleaseRepeat: pleaseRepeat,
	}
}

// Make a copy for re-sending the message.
// We set pleaseResend to false to avoid accidental loops.
func (msg *MsgNextLogIndex) AsResent() *MsgNextLogIndex {
	return &MsgNextLogIndex{
		BasicMessage: gpa.NewBasicMessage(msg.Recipient()),
		NextLogIndex: msg.NextLogIndex,
		Cause:        msg.Cause,
		PleaseRepeat: false,
	}
}

func (msg *MsgNextLogIndex) String() string {
	return fmt.Sprintf(
		"{MsgNextLogIndex[%v], sender=%v, nextLogIndex=%v, pleaseRepeat=%v",
		msg.Cause, msg.Sender().ShortString(), msg.NextLogIndex, msg.PleaseRepeat,
	)
}

func (msg *MsgNextLogIndex) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msgTypeNextLogIndex.ReadAndVerify(rr)
	msg.NextLogIndex = LogIndex(rr.ReadUint32())
	msg.Cause = MsgNextLogIndexCause(rr.ReadByte())
	msg.PleaseRepeat = rr.ReadBool()
	return rr.Err
}

func (msg *MsgNextLogIndex) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	msgTypeNextLogIndex.Write(ww)
	ww.WriteUint32(msg.NextLogIndex.AsUint32())
	ww.WriteByte(byte(msg.Cause))
	ww.WriteBool(msg.PleaseRepeat)
	return ww.Err
}
