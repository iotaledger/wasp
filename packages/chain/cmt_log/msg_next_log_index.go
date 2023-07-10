// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log

import (
	"fmt"
	"io"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type MsgNextLogIndexCause byte

const (
	MsgNextLogIndexCauseConsCompleted MsgNextLogIndexCause = iota
	MsgNextLogIndexCauseL1ReplacedAO
	MsgNextLogIndexCauseRecover
)

type MsgNextLogIndex struct {
	gpa.BasicMessage
	NextLogIndex LogIndex               // Proposal is to go to this LI without waiting for a consensus.
	NextBaseAO   *isc.AliasOutputWithID // Using this AO as a base.
	Cause        MsgNextLogIndexCause
	PleaseRepeat bool // If true, the receiver should resend its latest message back to the sender.
}

var _ gpa.Message = new(MsgNextLogIndex)

func NewMsgNextLogIndex(recipient gpa.NodeID, nextLogIndex LogIndex, nextBaseAO *isc.AliasOutputWithID, cause MsgNextLogIndexCause, pleaseRepeat bool) *MsgNextLogIndex {
	return &MsgNextLogIndex{
		BasicMessage: gpa.NewBasicMessage(recipient),
		NextLogIndex: nextLogIndex,
		NextBaseAO:   nextBaseAO,
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
		NextBaseAO:   msg.NextBaseAO,
		PleaseRepeat: false,
	}
}

func (msg *MsgNextLogIndex) String() string {
	return fmt.Sprintf(
		"{MsgNextLogIndex, sender=%v, nextLogIndex=%v, nextBaseAO=%v, pleaseRepeat=%v",
		msg.Sender().ShortString(), msg.NextLogIndex, msg.NextBaseAO, msg.PleaseRepeat,
	)
}

func (msg *MsgNextLogIndex) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msgTypeNextLogIndex.ReadAndVerify(rr)
	msg.NextLogIndex = LogIndex(rr.ReadUint32())
	msg.NextBaseAO = new(isc.AliasOutputWithID)
	rr.Read(msg.NextBaseAO)
	msg.Cause = MsgNextLogIndexCause(rr.ReadByte())
	msg.PleaseRepeat = rr.ReadBool()
	return rr.Err
}

func (msg *MsgNextLogIndex) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	msgTypeNextLogIndex.Write(ww)
	ww.WriteUint32(msg.NextLogIndex.AsUint32())
	ww.Write(msg.NextBaseAO)
	ww.WriteByte(byte(msg.Cause))
	ww.WriteBool(msg.PleaseRepeat)
	return ww.Err
}
