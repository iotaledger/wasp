// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type msgNextLogIndex struct {
	gpa.BasicMessage
	nextLogIndex LogIndex               // Proposal is to go to this LI without waiting for a consensus.
	nextBaseAO   *isc.AliasOutputWithID // Using this AO as a base.
	pleaseRepeat bool                   // If true, the receiver should resend its latest message back to the sender.
}

var _ gpa.Message = &msgNextLogIndex{}

func newMsgNextLogIndex(recipient gpa.NodeID, nextLogIndex LogIndex, nextBaseAO *isc.AliasOutputWithID, pleaseRepeat bool) *msgNextLogIndex {
	return &msgNextLogIndex{
		BasicMessage: gpa.NewBasicMessage(recipient),
		nextLogIndex: nextLogIndex,
		nextBaseAO:   nextBaseAO,
		pleaseRepeat: pleaseRepeat,
	}
}

// Make a copy for re-sending the message.
// We set pleaseResend to false to avoid accidental loops.
func (m *msgNextLogIndex) AsResent() *msgNextLogIndex {
	return &msgNextLogIndex{
		BasicMessage: gpa.NewBasicMessage(m.Recipient()),
		nextLogIndex: m.nextLogIndex,
		nextBaseAO:   m.nextBaseAO,
		pleaseRepeat: false,
	}
}

func (m *msgNextLogIndex) String() string {
	return fmt.Sprintf(
		"{msgNextLogIndex, sender=%v, nextLogIndex=%v, nextBaseAO=%v, pleaseRepeat=%v",
		m.Sender().ShortString(), m.nextLogIndex, m.nextBaseAO, m.pleaseRepeat,
	)
}

func (m *msgNextLogIndex) MarshalBinary() ([]byte, error) {
	w := &bytes.Buffer{}
	if err := rwutil.WriteByte(w, msgTypeNextLogIndex); err != nil {
		return nil, fmt.Errorf("cannot marshal type=msgTypeNextLogIndex: %w", err)
	}
	if err := rwutil.WriteUint32(w, m.nextLogIndex.AsUint32()); err != nil {
		return nil, fmt.Errorf("cannot marshal msgNextLogIndex.nextLogIndex: %w", err)
	}
	if err := rwutil.WriteBytes(w, m.nextBaseAO.Bytes()); err != nil {
		return nil, fmt.Errorf("cannot marshal msgNextLogIndex.nextBaseAO: %w", err)
	}
	if err := rwutil.WriteBool(w, m.pleaseRepeat); err != nil {
		return nil, fmt.Errorf("cannot marshal msgNextLogIndex.pleaseRepeat: %w", err)
	}
	return w.Bytes(), nil
}

func (m *msgNextLogIndex) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	msgType, err := rwutil.ReadByte(r)
	if err != nil {
		return err
	}
	if msgType != msgTypeNextLogIndex {
		return fmt.Errorf("unexpected msgType=%v in cmtLog.msgNextLogIndex", msgType)
	}
	var nextLogIndex uint32
	if nextLogIndex, err = rwutil.ReadUint32(r); err != nil {
		return fmt.Errorf("cannot unmarshal msgNextLogIndex.nextLogIndex: %w", err)
	}
	m.nextLogIndex = LogIndex(nextLogIndex)
	nextAOBin, err := rwutil.ReadBytes(r)
	if err != nil {
		return fmt.Errorf("cannot unmarshal msgNextLogIndex.nextBaseAO: %w", err)
	}
	m.nextBaseAO, err = isc.NewAliasOutputWithIDFromBytes(nextAOBin)
	if err != nil {
		return fmt.Errorf("cannot decode msgNextLogIndex.nextBaseAO: %w", err)
	}
	if m.pleaseRepeat, err = rwutil.ReadBool(r); err != nil {
		return fmt.Errorf("cannot unmarshal msgNextLogIndex.pleaseRepeat: %w", err)
	}
	return nil
}
