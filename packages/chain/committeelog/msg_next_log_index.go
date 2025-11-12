// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package committeelog

import (
	"fmt"
)

type MsgNextLogIndexCause byte

func (c MsgNextLogIndexCause) String() string {
	switch c {
	case MsgNextLogIndexCauseStarted:
		return "Started"
	default:
		return fmt.Sprintf("%v", byte(c))
	}
}

const (
	MsgNextLogIndexCauseStarted MsgNextLogIndexCause = iota // Consensus is started, maybe we have to catch up with it.
)

type MsgNextLogIndex struct {
	NextLogIndex LogIndex             // Proposal is to go to this LI without waiting for a consensus.
	Cause        MsgNextLogIndexCause // Reason for the proposal.
	PleaseRepeat bool                 // If true, the receiver should resend its latest message back to the sender.
}

func NewMsgNextLogIndex(nextLogIndex LogIndex, cause MsgNextLogIndexCause, pleaseRepeat bool) *MsgNextLogIndex {
	return &MsgNextLogIndex{
		NextLogIndex: nextLogIndex,
		Cause:        cause,
		PleaseRepeat: pleaseRepeat,
	}
}

// AsResent makes a copy for re-sending the message.
// We set pleaseResend to false to avoid accidental loops.
func (msg *MsgNextLogIndex) AsResent() *MsgNextLogIndex {
	return &MsgNextLogIndex{
		NextLogIndex: msg.NextLogIndex,
		Cause:        msg.Cause,
		PleaseRepeat: false,
	}
}

func (msg *MsgNextLogIndex) String() string {
	return fmt.Sprintf(
		"{MsgNextLogIndex[%v], nextLogIndex=%v, pleaseRepeat=%v",
		msg.Cause, msg.NextLogIndex, msg.PleaseRepeat,
	)
}
