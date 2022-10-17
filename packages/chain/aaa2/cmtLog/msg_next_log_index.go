// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"bytes"

	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util"
)

type msgNextLogIndex struct {
	gpa.BasicMessage
	nextLogIndex journal.LogIndex // Proposal is to go to this LI without waiting for a consensus.
}

var _ gpa.Message = &msgNextLogIndex{}

func newMsgNextLogIndex(recipient gpa.NodeID, nextLogIndex journal.LogIndex) *msgNextLogIndex {
	return &msgNextLogIndex{
		BasicMessage: gpa.NewBasicMessage(recipient),
		nextLogIndex: nextLogIndex,
	}
}

func (m *msgNextLogIndex) MarshalBinary() ([]byte, error) {
	w := &bytes.Buffer{}
	if err := util.WriteByte(w, msgTypeNextLogIndex); err != nil {
		return nil, xerrors.Errorf("cannot marshal type=msgTypeNextLogIndex: %w", err)
	}
	if err := util.WriteUint32(w, m.nextLogIndex.AsUint32()); err != nil {
		return nil, xerrors.Errorf("cannot marshal msgNextLogIndex.nextLogIndex: %w", err)
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
		return xerrors.Errorf("unexpected msgType=%v in cmtLog.msgNextLogIndex", msgType)
	}
	var nextLogIndex uint32
	if err := util.ReadUint32(r, &nextLogIndex); err != nil {
		return xerrors.Errorf("cannot unmarshal msgNextLogIndex.nextLogIndex: %w", err)
	}
	m.nextLogIndex = journal.LogIndex(nextLogIndex)
	return nil
}
