// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
)

type msgNextLogIndex struct {
	gpa.BasicMessage
	nextLogIndex LogIndex               // Proposal is to go to this LI without waiting for a consensus.
	nextBaseAO   *isc.AliasOutputWithID // Using this AO as a base.
}

var _ gpa.Message = &msgNextLogIndex{}

func newMsgNextLogIndex(recipient gpa.NodeID, nextLogIndex LogIndex, nextBaseAO *isc.AliasOutputWithID) *msgNextLogIndex {
	return &msgNextLogIndex{
		BasicMessage: gpa.NewBasicMessage(recipient),
		nextLogIndex: nextLogIndex,
		nextBaseAO:   nextBaseAO,
	}
}

func (m *msgNextLogIndex) String() string {
	return fmt.Sprintf("{msgNextLogIndex, sender=%v, nextLogIndex=%v, nextBaseAO=%v", m.Sender().ShortString(), m.nextLogIndex, m.nextBaseAO)
}

func (m *msgNextLogIndex) MarshalBinary() ([]byte, error) {
	w := &bytes.Buffer{}
	if err := util.WriteByte(w, msgTypeNextLogIndex); err != nil {
		return nil, fmt.Errorf("cannot marshal type=msgTypeNextLogIndex: %w", err)
	}
	if err := util.WriteUint32(w, m.nextLogIndex.AsUint32()); err != nil {
		return nil, fmt.Errorf("cannot marshal msgNextLogIndex.nextLogIndex: %w", err)
	}
	if err := util.WriteBytes16(w, m.nextBaseAO.Bytes()); err != nil {
		return nil, fmt.Errorf("cannot marshal msgNextLogIndex.nextBaseAO: %w", err)
	}
	return w.Bytes(), nil
}

func (m *msgNextLogIndex) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	msgType, err := util.ReadByte(r)
	if err != nil {
		return err
	}
	if msgType != msgTypeNextLogIndex {
		return fmt.Errorf("unexpected msgType=%v in cmtLog.msgNextLogIndex", msgType)
	}
	var nextLogIndex uint32
	if err := util.ReadUint32(r, &nextLogIndex); err != nil {
		return fmt.Errorf("cannot unmarshal msgNextLogIndex.nextLogIndex: %w", err)
	}
	m.nextLogIndex = LogIndex(nextLogIndex)
	nextAOBin, err := util.ReadBytes16(r)
	if err != nil {
		return fmt.Errorf("cannot unmarshal msgNextLogIndex.nextBaseAO: %w", err)
	}
	m.nextBaseAO, err = isc.NewAliasOutputWithIDFromBytes(nextAOBin)
	if err != nil {
		return fmt.Errorf("cannot decode msgNextLogIndex.nextBaseAO: %w", err)
	}
	return nil
}
